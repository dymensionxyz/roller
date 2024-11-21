package setup

import (
	"context"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
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

			relayerHome := relayerutils.GetHomeDir(home)
			relayerConfigPath := relayerutils.GetConfigFilePath(relayerHome)

			raID, hd, err := relayerutils.GetRollappToRunFor(home)
			if err != nil {
				pterm.Error.Println("failed to determine what RollApp to run for:", err)
				return
			}

			_, err = rollapputils.ValidateChainID(hd.ID)
			if err != nil {
				pterm.Error.Printf("'%s' is not a valid Hub ID: %v", raID, err)
				return
			}

			_, err = rollapputils.ValidateChainID(raID)
			if err != nil {
				pterm.Error.Printf("'%s' is not a valid RollApp ID: %v", raID, err)
				return
			}

			ok, err := rollapp.IsRegistered(raID, *hd)
			if err != nil {
				pterm.Error.Printf("'%s' is not a valid RollApp ID: %v", raID, err)
				return
			}

			if !ok {
				pterm.Error.Printf("%s rollapp not registered on %s", raID, hd.ID)
				return
			}

			raRpc, err := sequencerutils.GetRpcEndpointFromChain(raID, *hd)
			if err != nil {
				pterm.Error.Println("failed to retrieve rollapp rpc endpoint: ", err)
				return
			}

			raData := consts.RollappData{
				ID:     raID,
				RpcUrl: fmt.Sprintf("%s:%d", raRpc, 443),
			}
			relayerLogFilePath := logging.GetRelayerLogPath(home)
			relayerLogger := logging.GetLogger(relayerLogFilePath)

			rollappChainData, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
				home,
				raID,
				*hd,
			)
			errorhandling.PrettifyErrorIfExists(err)

			err = rollappChainData.ValidateConfig()
			if err != nil {
				pterm.Error.Println("rollapp data validation error: ", err)
				return
			}
			pterm.Info.Println("rollapp chain data validation passed")

			raResp, err := rollapp.GetMetadataFromChain(raID, *hd)
			if err != nil {
				pterm.Error.Println("failed to fetch rollapp information from hub: ", err)
				return
			}
			err = genesis.DownloadGenesis(home, raResp.Rollapp.Metadata.GenesisUrl)
			if err != nil {
				pterm.Error.Println("failed to download genesis file: ", err)
				return
			}
			as, err := genesis.GetGenesisAppState(home)
			if err != nil {
				pterm.Error.Println("failed to get genesis app state: ", err)
				return
			}
			ctx := context.Background()
			conf := &firebase.Config{ProjectID: "drs-metadata"}
			app, err := firebase.NewApp(ctx, conf, option.WithoutAuthentication())
			if err != nil {
				pterm.Error.Printfln("failed to initialize firebase app: %v", err)
				return
			}
			drsVersion := strconv.Itoa(as.RollappParams.Params.DrsVersion)

			client, err := app.Firestore(ctx)
			if err != nil {
				pterm.Error.Printfln("failed to create firestore client: %v", err)
				return
			}
			defer client.Close()

			// Fetch DRS version information using the nested collection path
			// Path format: versions/{version}/revisions/{revision}
			drsDoc := client.Collection("versions").
				Doc(drsVersion).
				Collection("revisions").
				OrderBy("timestamp", firestore.Desc).
				Limit(1).
				Documents(ctx)

			doc, err := drsDoc.Next()
			if err == iterator.Done {
				pterm.Error.Printfln("DRS version not found for %s", drsVersion)
				return
			}
			if err != nil {
				pterm.Error.Printfln("DRS version not found for %s", drsVersion)
				return
			}

			var drsInfo dependencies.DrsVersionInfo
			if err := doc.DataTo(&drsInfo); err != nil {
				pterm.Error.Printfln("DRS version not found for %s", drsVersion)
				return
			}

			dep := dependencies.DefaultRelayerPrebuiltDependencies()
			for _, v := range dep {
				err := dependencies.InstallBinaryFromRelease(v)
				if err != nil {
					pterm.Error.Printfln("failed to install binary: %s", err)
					return
				}
			}

			rbi := dependencies.NewRollappBinaryInfo(
				raResp.Rollapp.GenesisInfo.Bech32Prefix,
				drsInfo.Commit,
				strings.ToLower(raResp.Rollapp.VmType),
			)

			raDep := dependencies.DefaultRollappDependency(rbi)
			err = dependencies.InstallBinaryFromRepo(raDep, raDep.DependencyName)
			if err != nil {
				pterm.Error.Printfln("failed to install binary: %s", err)
				return
			}

			// things to check:
			// 1. relayer folder exists
			isRelayerDirPresent, err := filesystem.DirNotEmpty(relayerHome)
			if err != nil {
				pterm.Error.Printf("failed to check %s: %v\n", relayerHome, err)
				return
			}

			var ibcPathChains *relayerutils.IbcPathChains

			if isRelayerDirPresent {
				ibcPathChains, err = relayerutils.ValidateIbcPathChains(
					relayerHome,
					raID,
					*hd,
				)
				if err != nil {
					pterm.Error.Printf(
						"validate relayer config IBC path %s: %v\n",
						relayerHome,
						err,
					)
					return
				}
			} else {
				err = os.MkdirAll(relayerHome, 0o755)
				if err != nil {
					pterm.Error.Printf("failed to create %s: %v\n", relayerHome, err)
					return
				}
			}

			if ibcPathChains != nil {
				if !ibcPathChains.DefaultPathOk || !ibcPathChains.SrcChainOk ||
					!ibcPathChains.DstChainOk {
					pterm.Warning.Println("relayer config verification failed...")
					if ibcPathChains.DefaultPathOk {
						pterm.Info.Printfln(
							"removing path from config %s",
							consts.DefaultRelayerPath,
						)
						err := relayer.DeletePath(*rollappChainData)
						if err != nil {
							pterm.Error.Printf("failed to delete relayer IBC path: %v\n", err)
							return
						}
					}

					pterm.Info.Println("populating relayer config with correct values...")
					err = relayerutils.InitializeRelayer(home, *rollappChainData)
					if err != nil {
						pterm.Error.Printf("failed to initialize relayer config: %v\n", err)
						return
					}

					if err := relayer.CreatePath(*rollappChainData); err != nil {
						pterm.Error.Printf("failed to create relayer IBC path: %v\n", err)
						return
					}
				}
			} else {
				pterm.Info.Println("populating relayer config with correct values...")
				err = relayerutils.InitializeRelayer(home, *rollappChainData)
				if err != nil {
					pterm.Error.Printf("failed to initialize relayer config: %v\n", err)
					return
				}

				if err := relayer.CreatePath(*rollappChainData); err != nil {
					pterm.Error.Printf("failed to create relayer IBC path: %v\n", err)
					return
				}
			}

			if err := relayerutils.UpdateConfigWithDefaultValues(relayerHome, *rollappChainData); err != nil {
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
			rly := relayer.NewRelayer(
				home,
				raData.ID,
				hd.ID,
			)
			rly.SetLogger(relayerLogger)
			logFileOption := logging.WithLoggerLogging(relayerLogger)

			srcIbcChannel, dstIbcChannel, err := rly.LoadActiveChannel(raData, *hd)
			if err != nil {
				pterm.Error.Printf("failed to load active channel, %v", err)
				return
			}

			if srcIbcChannel != "" && dstIbcChannel != "" {
				pterm.Info.Println("updating application relayer config")

				rollappIbcConnection, hubIbcConnection, err := rly.GetActiveConnections(
					raData,
					*hd,
				)
				if err != nil {
					pterm.Error.Printf("failed to retrieve active connections: %v\n", err)
					return
				}

				updates := map[string]interface{}{
					// hub
					fmt.Sprintf("paths.%s.src.client-id", consts.DefaultRelayerPath):     hubIbcConnection.ClientID,
					fmt.Sprintf("paths.%s.src.connection-id", consts.DefaultRelayerPath): hubIbcConnection.ID,

					// ra
					fmt.Sprintf("paths.%s.dst.client-id", consts.DefaultRelayerPath):     rollappIbcConnection.ClientID,
					fmt.Sprintf("paths.%s.dst.connection-id", consts.DefaultRelayerPath): rollappIbcConnection.ID,
				}
				err = yamlconfig.UpdateNestedYAML(relayerConfigPath, updates)
				if err != nil {
					pterm.Error.Printf("Error updating YAML: %v\n", err)
					return
				}

				pterm.Info.Println("existing IBC channels found ")
				pterm.Info.Println("Hub: ", srcIbcChannel)
				pterm.Info.Println("RollApp: ", dstIbcChannel)
				return
			}

			defer func() {
				pterm.Info.Println("reverting dymint config to 1h")
				err := dymintutils.UpdateDymintConfigForIBC(home, "1h0m0s", true)
				if err != nil {
					pterm.Error.Println("failed to update dymint config: ", err)
					return
				}
			}()

			canCreateIbcConnectionOnCurrentNode, err := relayerutils.NewIbcConnenctionCanBeCreatedOnCurrentNode(
				home,
				raID,
			)
			if err != nil {
				pterm.Error.Println(
					"failed to determine whether connection can be created from this node:",
					err,
				)
				return
			}

			if !canCreateIbcConnectionOnCurrentNode {
				pterm.Error.Println(err)
				return
			}

			err = relayerutils.InitializeRelayer(home, *rollappChainData)
			if err != nil {
				pterm.Error.Println("failed to initialize relayer:", err)
				return
			}

			pterm.Info.Println("let's create that IBC connection, shall we?")
			seq := sequencer.GetInstance(*rollappChainData)

			health := fmt.Sprintf(consts.DefaultRollappRPC+"%s", "/health")
			// dymintutils.WaitForHealthyRollApp("http://localhost:26657/health")
			dymintutils.WaitForHealthyRollApp(health)
			err = relayer.WaitForValidRollappHeight(seq)
			if err != nil {
				pterm.Error.Printf("rollapp did not reach valid height: %v\n", err)
				return
			}

			// add whitelisted relayers
			seqAddr, err := sequencerutils.GetSequencerAccountAddress(*rollappChainData)
			if err != nil {
				pterm.Error.Printf("failed to get sequencer address: %v\n", err)
				return
			}

			isRlyKeyWhitelisted, err := relayerutils.IsRelayerRollappKeyWhitelisted(
				seqAddr,
				relKeys[consts.KeysIds.RollappRelayer].Address,
				*hd,
			)
			if err != nil {
				pterm.Error.Printf("failed to check if relayer key is whitelisted: %v\n", err)
				return
			}

			if !isRlyKeyWhitelisted {
				pterm.Warning.Printfln(
					"relayer key (%s) is not whitelisted, updating whitelisted relayers",
					relKeys[consts.KeysIds.RollappRelayer].Address,
				)

				err := sequencerutils.UpdateWhitelistedRelayers(
					home,
					relKeys[consts.KeysIds.RollappRelayer].Address,
					*hd,
				)
				if err != nil {
					pterm.Error.Println("failed to update whitelisted relayers:", err)
					return
				}
			}

			raOpAddr, err := sequencerutils.GetSequencerOperatorAddress(home)
			if err != nil {
				pterm.Error.Println("failed to get RollApp's operator address:", err)
				return
			}

			wrSpinner, _ := pterm.DefaultSpinner.Start(
				"waiting for the whitelisted relayer to propagate to RollApp (this might take a while)",
			)
			for {
				r, err := sequencerutils.GetWhitelistedRelayersOnRa(raOpAddr)
				if err != nil {
					pterm.Error.Println("failed to get whitelisted relayers:", err)
					return
				}

				if len(r) == 0 &&
					slices.Contains(r, relKeys[consts.KeysIds.RollappRelayer].Address) {
					wrSpinner.UpdateText(
						"waiting for the whitelisted relayer to propagate to RollApp...",
					)
					time.Sleep(time.Second * 5)
					continue
				} else {
					// nolint: errcheck
					wrSpinner.Success("relayer whitelisted and propagated to rollapp")
					break
				}
			}

			pterm.Info.Println("setting block time to 5s for esstablishing IBC connection")
			err = dymintutils.UpdateDymintConfigForIBC(home, "5s", true)
			if err != nil {
				pterm.Error.Println("failed to update dymint config: ", err)
				return
			}

			if rly.ChannelReady() {
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

			if createIbcChannels {
				err = relayerutils.VerifyRelayerBalances(*hd)
				if err != nil {
					pterm.Error.Printf("failed to verify relayer balances: %v\n", err)
					return
				}

				pterm.Info.Println("establishing IBC transfer channel")
				channels, err := rly.CreateIBCChannel(
					logFileOption,
					raData,
					*hd,
				)
				if err != nil {
					pterm.Error.Printf("failed to create IBC channel: %v\n", err)
					return
				}

				srcIbcChannel = channels.Src
				dstIbcChannel = channels.Dst

				status := fmt.Sprintf(
					"Active\nrollapp: %s\n<->\nhub: %s",
					srcIbcChannel,
					dstIbcChannel,
				)

				pterm.Info.Println(status)
				err = rly.WriteRelayerStatus(status)
				if err != nil {
					fmt.Println(err)
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

	return relayerStartCmd
}
