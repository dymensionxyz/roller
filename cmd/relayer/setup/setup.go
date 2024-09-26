package setup

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	configutils "github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

// TODO: Test relaying on 35-C and update the prices
const (
	flagOverride = "override"
)

func Cmd() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "setup",
		Short: "Setup IBC connection between the Dymension hub and the RollApp.",
		Run: func(cmd *cobra.Command, args []string) {
			// TODO: there are too many things set here, might be worth to refactor
			home, _ := filesystem.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			relayerHome := filepath.Join(home, consts.ConfigDirName.Relayer)

			// check for roller config, if it's present - fetch the rollapp ID from there
			var raID string
			var env string
			var hd consts.HubData
			var runForExisting bool
			var rollerData configutils.RollappConfig

			rollerConfigFilePath := filepath.Join(home, consts.RollerConfigFileName)

			// fetch rollapp metadata from chain
			// retrieve rpc endpoint
			_, err := os.Stat(rollerConfigFilePath)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					pterm.Info.Println("existing roller configuration not found")
					runForExisting = false
				} else {
					pterm.Error.Println("failed to check existing roller config")
					return
				}
			} else {
				pterm.Info.Println("existing roller configuration found, retrieving RollApp ID from it")
				rollerData, err = tomlconfig.LoadRollerConfig(home)
				if err != nil {
					pterm.Error.Printf("failed to load rollapp config: %v\n", err)
					return
				}
				rollerRaID := rollerData.RollappID
				rollerHubData := rollerData.HubData
				msg := fmt.Sprintf(
					"the retrieved rollapp ID is: %s, would you like to initialize the relayer for this rollapp?",
					rollerRaID,
				)
				rlyFromRoller, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(msg).Show()
				if rlyFromRoller {
					raID = rollerRaID
					hd = rollerHubData
					runForExisting = true
				}

				if !rlyFromRoller {
					runForExisting = false
				}
			}

			if !runForExisting {
				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText("Please enter the RollApp ID").
					Show()
			}

			_, err = rollapputils.ValidateChainID(raID)
			if err != nil {
				pterm.Error.Printf("'%s' is not a valid RollApp ID: %v", raID, err)
				return
			}

			if !runForExisting {
				envs := []string{"playground"}
				env, _ = pterm.DefaultInteractiveSelect.
					WithDefaultText(
						"select the environment you want to initialize relayer for",
					).
					WithOptions(envs).
					Show()

				hd = consts.Hubs[env]
			}

			// retrieve rollapp rpc endpoints
			raRpc, err := sequencerutils.GetRpcEndpointFromChain(raID, hd)
			if err != nil {
				return
			}

			// check if there are active channels created for the rollapp
			relayerLogFilePath := utils.GetRelayerLogPath(home)
			relayerLogger := utils.GetLogger(relayerLogFilePath)

			raData := consts.RollappData{
				ID:     raID,
				RpcUrl: fmt.Sprintf("%s:%d", raRpc, 443),
			}

			rly := relayer.NewRelayer(
				home,
				raData.ID,
				hd.ID,
			)
			rly.SetLogger(relayerLogger)
			logFileOption := utils.WithLoggerLogging(relayerLogger)

			srcIbcChannel, dstIbcChannel, err := rly.LoadActiveChannel(raData, hd)
			if err != nil {
				pterm.Error.Printf("failed to load active channel, %v", err)
				return
			}

			if srcIbcChannel == "" || dstIbcChannel == "" {
				if !runForExisting {
					pterm.Error.Println(
						"existing channels not found, initial IBC setup must be run on a sequencer node",
					)
					return
				}

				if runForExisting && rollerData.NodeType != consts.NodeType.Sequencer {
					pterm.Error.Println(
						"existing channels not found, initial IBC setup must be run on a sequencer node",
					)
					return
				}

				pterm.Info.Println("let's create that IBC connection, shall we?")
			}

			fmt.Println("hub channel: ", srcIbcChannel)
			fmt.Println("ra channel: ", dstIbcChannel)
			// if yes -
			// add it to the relayer configuration
			// if not rpc endpoints are available - exit
			// if no -
			// prompt to create new ibc connection between hub and rollapp
			// create it, display the channel info
			return

			// you can't create the channels if the relayer is not running on the same host
			defer func() {
				pterm.Info.Println("reverting dymint config to 1h")
				err := dymintutils.UpdateDymintConfigForIBC(home, "1h0m0s", true)
				if err != nil {
					pterm.Error.Println("failed to update dymint config: ", err)
					return
				}
			}()

			// otherwise prompt
			as, err := genesisutils.GetGenesisAppState(home)
			if err != nil {
				pterm.Error.Printf("failed to get genesis app state: %v\n", err)
				return
			}
			rollappDenom := as.Bank.Supply[0].Denom

			err = globalutils.UpdateFieldInToml(rollerConfigFilePath, "base_denom", rollappDenom)
			if err != nil {
				pterm.Error.Println("failed to set base denom in roller.toml")
				return
			}

			seq := sequencer.GetInstance(rollerData)

			rollappChainData, err := tomlconfig.LoadRollappMetadataFromChain(
				home,
				rollerData.RollappID,
				&rollerData.HubData,
				string(rollerData.VMType),
			)
			errorhandling.PrettifyErrorIfExists(err)

			// check if there are active channels created for the rollapp
			// if yes -
			// fetch rollapp metadata from chain
			// retrieve rpc endpoint
			// add it to the relayer configuration
			// if not rpc endpoints are available - exit
			// if no -
			// prompt to create new ibc connection between hub and rollapp
			// create it, display the channel info

			/* ---------------------------- Initialize relayer --------------------------- */

			dymintutils.WaitForHealthyRollApp("http://localhost:26657/health")
			err = relayer.WaitForValidRollappHeight(seq)
			if err != nil {
				pterm.Error.Printf("rollapp did not reach valid height: %v\n", err)
				return
			}
			outputHandler := initconfig.NewOutputHandler(false)
			isRelayerInitialized, err := filesystem.DirNotEmpty(relayerHome)
			if err != nil {
				pterm.Error.Printf("failed to check %s: %v\n", relayerHome, err)
				return
			}

			var shouldOverwrite bool
			if isRelayerInitialized {
				outputHandler.StopSpinner()
				shouldOverwrite, err = outputHandler.PromptOverwriteConfig(relayerHome)
				if err != nil {
					pterm.Error.Printf("failed to get your input: %v\n", err)
					return
				}
			}

			if shouldOverwrite {
				pterm.Info.Println("overriding the existing relayer configuration")
				err = os.RemoveAll(relayerHome)
				if err != nil {
					pterm.Error.Printf("failed to recuresively remove %s: %v\n", relayerHome, err)
					return
				}

				err := servicemanager.RemoveServiceFiles(consts.RelayerSystemdServices)
				if err != nil {
					pterm.Error.Printf("failed to remove relayer systemd services: %v\n", err)
					return
				}

				err = os.MkdirAll(relayerHome, 0o755)
				if err != nil {
					pterm.Error.Printf("failed to create %s: %v\n", relayerHome, err)
					return
				}
			}

			if !isRelayerInitialized || shouldOverwrite {
				// preflight checks
				blockInformation, err := rollapputils.GetCurrentHeight()
				if err != nil {
					pterm.Error.Printf("failed to get current block height: %v\n", err)
					return
				}
				currentHeight, err := strconv.Atoi(
					blockInformation.Block.Header.Height,
				)
				if err != nil {
					pterm.Error.Printf("failed to get current block height: %v\n", err)
					return
				}

				if currentHeight <= 2 {
					pterm.Warning.Println("current height is too low, updating dymint config")
					err = dymintutils.UpdateDymintConfigForIBC(home, "5s", false)
					if err != nil {
						pterm.Error.Println("failed to update dymint config: ", err)
						return
					}
				}

				rollappPrefix := rollappChainData.Bech32Prefix
				if err != nil {
					pterm.Error.Printf("failed to retrieve bech32_prefix: %v\n", err)
					return
				}

				pterm.Info.Println("initializing relayer config")
				err = initconfig.InitializeRelayerConfig(
					relayer.ChainConfig{
						ID:            rollerData.RollappID,
						RPC:           consts.DefaultRollappRPC,
						Denom:         rollappDenom,
						AddressPrefix: rollappPrefix,
						GasPrices:     "2000000000",
					}, relayer.ChainConfig{
						ID:            rollerData.HubData.ID,
						RPC:           rollerData.HubData.RPC_URL,
						Denom:         consts.Denoms.Hub,
						AddressPrefix: consts.AddressPrefixes.Hub,
						GasPrices:     rollerData.HubData.GAS_PRICE,
					}, rollerData,
				)
				if err != nil {
					pterm.Error.Printf(
						"failed to initialize relayer config: %v\n",
						err,
					)
					return
				}

				keys, err := initconfig.GenerateRelayerKeys(rollerData)
				if err != nil {
					pterm.Error.Printf("failed to create relayer keys: %v\n", err)
					return
				}

				for _, key := range keys {
					key.Print(utils.WithMnemonic(), utils.WithName())
				}

				keysToFund, err := initconfig.GetRelayerKeys(rollerData)
				pterm.Info.Println("please fund the hub relayer key with at least 20 dym tokens: ")
				for _, k := range keysToFund {
					k.Print(utils.WithName())
				}
				proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
					WithDefaultText(
						"press 'y' when the wallets are funded",
					).Show()
				if !proceed {
					return
				}

				if err != nil {
					pterm.Error.Printf("failed to create relayer keys: %v\n", err)
					return
				}

				if err := relayer.CreatePath(rollerData); err != nil {
					pterm.Error.Printf("failed to create relayer IBC path: %v\n", err)
					return
				}

				pterm.Info.Println("updating application relayer config")
				relayerConfigPath := filepath.Join(relayerHome, "config", "config.yaml")
				updates := map[string]interface{}{
					fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.HubData.ID): 1.5,
					fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.RollappID):  1.3,
					fmt.Sprintf("chains.%s.value.is-dym-hub", rollerData.HubData.ID):     true,
					fmt.Sprintf(
						"chains.%s.value.http-addr",
						rollerData.HubData.ID,
					): rollerData.HubData.API_URL,
					fmt.Sprintf("chains.%s.value.is-dym-rollapp", rollerData.RollappID): true,
					"extra-codecs": []string{
						"ethermint",
					},
				}
				err = yamlconfig.UpdateNestedYAML(relayerConfigPath, updates)
				if err != nil {
					pterm.Error.Printf("Error updating YAML: %v\n", err)
					return
				}

				err = dymintutils.UpdateDymintConfigForIBC(home, "5s", false)
				if err != nil {
					pterm.Error.Printf("Error updating YAML: %v\n", err)
					return
				}
			}

			if isRelayerInitialized && !shouldOverwrite {
				pterm.Info.Println("ensuring relayer keys are present")
				kc := initconfig.GetRelayerKeysConfig(rollerData)

				for k, v := range kc {
					pterm.Info.Printf("checking %s\n", k)

					switch v.ID {
					case consts.KeysIds.RollappRelayer:
						chainId := rollerData.RollappID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}

						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollerData.RollappID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					case consts.KeysIds.HubRelayer:
						chainId := rollerData.HubData.ID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}
						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollerData.HubData.ID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					default:
						pterm.Error.Println("invalid key name", err)
						return
					}
				}
			}

			if isRelayerInitialized && !shouldOverwrite {
				pterm.Info.Println("ensuring relayer keys are present")
				kc := initconfig.GetRelayerKeysConfig(rollerData)

				for k, v := range kc {
					pterm.Info.Printf("checking %s\n", k)

					switch v.ID {
					case consts.KeysIds.RollappRelayer:
						chainId := rollerData.RollappID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}

						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollerData.RollappID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					case consts.KeysIds.HubRelayer:
						chainId := rollerData.HubData.ID
						isPresent, err := utils.IsRlyAddressWithNameInKeyring(v, chainId)
						if err != nil {
							pterm.Error.Printf("failed to check address: %v\n", err)
							return
						}
						if !isPresent {
							key, err := initconfig.AddRlyKey(v, rollerData.HubData.ID)
							if err != nil {
								pterm.Error.Printf("failed to add key: %v\n", err)
							}

							key.Print(utils.WithMnemonic(), utils.WithName())
						}
					default:
						pterm.Error.Println("invalid key name", err)
						return
					}
				}
			}

			err = verifyRelayerBalances(rollerData)
			if err != nil {
				return
			}

			// errorhandling.RequireMigrateIfNeeded(rollappConfig)

			err = rollerData.Validate()
			if err != nil {
				pterm.Error.Printf("failed to validate rollapp config: %v\n", err)
				return
			}

			var createIbcChannels bool

			if rly.ChannelReady() && !shouldOverwrite {
				pterm.DefaultSection.WithIndentCharacter("💈").
					Println("IBC transfer channel is already established!")

				status := fmt.Sprintf(
					"Active\nrollapp: %s\n<->\nhub: %s",
					rly.SrcChannel,
					rly.DstChannel,
				)
				err := rly.WriteRelayerStatus(status)
				if err != nil {
					fmt.Println(err)
					return
				}

				pterm.Info.Println(status)
				return
			}

			if !rly.ChannelReady() {
				createIbcChannels, _ = pterm.DefaultInteractiveConfirm.WithDefaultText(
					fmt.Sprintf(
						"no channel found. would you like to create a new IBC channel for %s?",
						rollerData.RollappID,
					),
				).Show()

				if !createIbcChannels {
					pterm.Warning.Println("you can't run a relayer without an ibc channel")
					return
				}
			}

			// TODO: look up relayer keys
			if createIbcChannels || shouldOverwrite {
				err = verifyRelayerBalances(rollerData)
				if err != nil {
					pterm.Error.Printf("failed to verify relayer balances: %v\n", err)
					return
				}

				pterm.Info.Println("establishing IBC transfer channel")
				seq := sequencer.GetInstance(rollerData)
				if seq == nil {
					pterm.Error.Println("failed to get sequencer sequencer instance")
					return
				}

				_, err = rly.CreateIBCChannel(shouldOverwrite, logFileOption, raData, hd)
				if err != nil {
					pterm.Error.Printf("failed to create IBC channel: %v\n", err)
					return
				}
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

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}

func verifyRelayerBalances(rolCfg configutils.RollappConfig) error {
	insufficientBalances, err := relayer.GetRelayerInsufficientBalances(rolCfg)
	if err != nil {
		return err
	}
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)

	return nil
}
