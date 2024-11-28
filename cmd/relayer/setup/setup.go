package setup

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	firebaseutils "github.com/dymensionxyz/roller/utils/firebase"
	"github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/logging"
	relayerutils "github.com/dymensionxyz/roller/utils/relayer"
	"github.com/dymensionxyz/roller/utils/rollapp"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: cleanup required, a lot of duplicate code in this cmd
func Cmd() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup IBC connection between the Dymension hub and the RollApp.",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: there are too many things set here, might be worth to refactor
			home, _ := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)

			raData, hd, kb, err := getPreRunInfo(home)
			if err != nil {
				pterm.Error.Println("failed to run pre-flight checks: ", err)
				return
			}
			rly := relayer.NewRelayer(
				home,
				*raData,
				*hd,
			)
			relayerLogFilePath := logging.GetRelayerLogPath(home)
			relayerLogger := logging.GetLogger(relayerLogFilePath)
			rly.SetLogger(relayerLogger)

			rollappChainData, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
				home,
				raData.ID,
				*hd,
			)
			errorhandling.PrettifyErrorIfExists(err)
			rollappChainData.KeyringBackend = consts.SupportedKeyringBackend(kb)

			err = rollappChainData.ValidateConfig()
			if err != nil {
				pterm.Error.Println("rollapp data validation error: ", err)
				return
			}
			pterm.Info.Println("rollapp chain data validation passed")

			err = installRelayerDependencies(home, rly.Rollapp.ID, *hd)
			if err != nil {
				pterm.Error.Println("failed to install relayer dependencies: ", err)
				return
			}

			// things to check:
			// 1. relayer folder exists
			dirExist, err := filesystem.DirNotEmpty(rly.RelayerHome)
			if err != nil {
				pterm.Error.Printf("failed to check %s: %v\n", rly.RelayerHome, err)
				return
			}

			if !dirExist {
				err = os.MkdirAll(rly.RelayerHome, 0o755)
				if err != nil {
					pterm.Error.Printf("failed to create %s: %v\n", rly.RelayerHome, err)
					return
				}
			}

			pterm.Info.Println("populating relayer config with correct values...")
			err = relayerutils.InitializeRelayer(home, *rollappChainData)
			if err != nil {
				pterm.Error.Printf("failed to initialize relayer config: %v\n", err)
				return
			}

			var rlyCfg relayer.Config
			err = rlyCfg.Load(rly.ConfigFilePath)
			if err != nil {
				pterm.Error.Println("failed to load relayer config: ", err)
				return
			}

			pterm.Info.Println("verifying path in relayer config")
			if rlyCfg.GetPath() == nil {
				pterm.Error.Println("no existing path")
				if err := rlyCfg.CreatePath(*rollappChainData); err != nil {
					pterm.Error.Printf("failed to create relayer IBC path: %v\n", err)
					return
				}
			}

			if err := rly.UpdateConfigWithDefaultValues(*rollappChainData); err != nil {
				pterm.Error.Printf("failed to update relayer config file: %v\n", err)
				return
			}

			relKeys, err := relayerutils.EnsureKeysArePresentAndFunded(*rollappChainData)
			if err != nil {
				pterm.Error.Println(
					"failed to ensure relayer keys are created/funded:",
					err,
				)
				return
			}

			// 5. Are there existing channels ( from chain )
			logFileOption := logging.WithLoggerLogging(relayerLogger)
			err = rly.LoadActiveChannel(*raData, *hd)
			if err != nil {
				if errors.Is(err, relayer.ErrNoOpenChannel) {

					pterm.Warning.Println("No open channel found")
					var createIbcChannels bool
					if !rly.ChannelReady() {
						createIbcChannels, _ = pterm.DefaultInteractiveConfirm.WithDefaultText(
							fmt.Sprintf(
								"no channel found. would you like to create a new IBC channel for %s?",
								pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
									Sprint(rollappChainData.RollappID),
							),
						).Show()

						if !createIbcChannels {
							pterm.Warning.Println("you can't run a relayer without an ibc channel")
							return
						}
					}

					pterm.Info.Println("checking whether node is eligible to create ibc connection")
					canCreateIbc, err := relayerutils.NewIbcConnenctionCanBeCreatedOnCurrentNode(
						home,
						rly.Rollapp.ID,
					)
					if err != nil {
						pterm.Error.Println(
							"failed to determine whether connection can be created from this node:",
							err,
						)
						return
					}
					if !canCreateIbc {
						pterm.Error.Println(err)
						return
					}

					pterm.Info.Println("creating ibc connection")
					err = rly.HandleWhitelisting(
						relKeys[consts.KeysIds.RollappRelayer].Address,
						rollappChainData,
					)
					if err != nil {
						pterm.Error.Println("failed to handle whitelisting: ", err)
						return
					}

					err = rly.HandleIbcChannelCreation(home, *rollappChainData, logFileOption)
					if err != nil {
						pterm.Error.Println("failed to handle ibc channel creation: ", err)
						return
					}
				} else {
					pterm.Error.Printf("failed to load active channel, %v", err)
					return
				}
			}

			if rly.SrcChannel != "" && rly.DstChannel != "" {
				pterm.Info.Println("existing IBC channels found ")
				pterm.Info.Println("Hub chan: ", rly.SrcChannel)
				pterm.Info.Println("RollApp chan: ", rly.DstChannel)
				pterm.Info.Println("RollApp conn: ", rly.DstConnectionID)

				err := rly.ConnectionInfoFromRaConnID(*raData, rly.DstConnectionID)
				if err != nil {
					pterm.Error.Println("failed to get hub ibc connection: ", err)
					return
				}

				err = rly.UpdateDefaultPath()
				if err != nil {
					pterm.Error.Println("failed to update relayer config: ", err)
					return
				}
				return
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s load the necessary systemd services\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller relayer services load"),
				)
			}()
		},
	}

	return relayerStartCmd
}

func getPreRunInfo(home string) (*consts.RollappData, *consts.HubData, string, error) {
	raID, hd, kb, err := relayerutils.GetRollappToRunFor(home)
	if err != nil {
		pterm.Error.Println("failed to determine what RollApp to run for:", err)
		return nil, nil, "", err
	}

	_, err = rollapputils.ValidateChainID(hd.ID)
	if err != nil {
		pterm.Error.Printf("'%s' is not a valid Hub ID: %v", raID, err)
		return nil, nil, "", err
	}

	_, err = rollapputils.ValidateChainID(raID)
	if err != nil {
		pterm.Error.Printf("'%s' is not a valid RollApp ID: %v", raID, err)
		return nil, nil, "", err
	}

	ok, err := rollapp.IsRegistered(raID, *hd)
	if err != nil {
		pterm.Error.Printf("'%s' is not a valid RollApp ID: %v", raID, err)
		return nil, nil, "", err
	}

	if !ok {
		pterm.Error.Printf("%s rollapp not registered on %s", raID, hd.ID)
		return nil, nil, "", err
	}

	raRpc, err := sequencerutils.GetRpcEndpointFromChain(raID, *hd)
	if err != nil {
		pterm.Error.Println("failed to retrieve rollapp rpc endpoint: ", err)
		return nil, nil, "", err
	}

	raData := consts.RollappData{
		ID:     raID,
		RpcUrl: fmt.Sprintf("%s:%d", raRpc, 443),
	}
	return &raData, hd, kb, nil
}

func getDrsVersionFromGenesis(
	home string,
	raResp *rollapputils.ShowRollappResponse,
) (string, error) {
	err := genesis.DownloadGenesis(home, raResp.Rollapp.Metadata.GenesisUrl)
	if err != nil {
		return "", err
	}

	as, err := genesis.GetGenesisAppState(home)
	if err != nil {
		pterm.Error.Println("failed to get genesis app state: ", err)
		return "", err
	}
	drsVersion := strconv.Itoa(as.RollappParams.Params.DrsVersion)

	return drsVersion, nil
}

func installRelayerDependencies(
	home string,
	raID string,
	hd consts.HubData,
) error {
	raResp, err := rollapp.GetMetadataFromChain(raID, hd)
	if err != nil {
		return err
	}

	drsVersion, err := getDrsVersionFromGenesis(home, raResp)
	if err != nil {
		pterm.Error.Println("failed to get drs version from genesis: ", err)
		return err
	}

	drsInfo, err := firebaseutils.GetLatestDrsVersionCommit(drsVersion)
	if err != nil {
		pterm.Error.Println("failed to retrieve latest DRS version: ", err)
		return err
	}

	rbi := dependencies.NewRollappBinaryInfo(
		raResp.Rollapp.GenesisInfo.Bech32Prefix,
		drsInfo.Commit,
		strings.ToLower(raResp.Rollapp.VmType),
	)

	raDep := dependencies.DefaultRollappDependency(rbi)
	err = dependencies.InstallBinaryFromRepo(raDep, raDep.DependencyName)
	if err != nil {
		return err
	}

	rlyDep := dependencies.DefaultRelayerPrebuiltDependencies()
	err = dependencies.InstallBinaryFromRelease(rlyDep["rly"])
	if err != nil {
		return err
	}

	dymdDep := dependencies.DefaultDymdDependency()
	err = dependencies.InstallBinaryFromRelease(dymdDep)
	if err != nil {
		return err
	}

	return nil
}
