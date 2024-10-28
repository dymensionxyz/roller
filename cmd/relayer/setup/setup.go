package setup

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/logging"
	relayerutils "github.com/dymensionxyz/roller/utils/relayer"
	"github.com/dymensionxyz/roller/utils/rollapp"
	rollapputils "github.com/dymensionxyz/roller/utils/rollapp"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: Test relaying on 35-C and update the prices
const (
	flagOverride = "override"
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

			ok, err := rollapp.IsRollappRegistered(raID, *hd)
			if err != nil {
				pterm.Error.Printf("'%s' is not a valid RollApp ID: %v", raID, err)
				return
			}

			if !ok {
				pterm.Error.Printf("%s rollapp not registered on %s", raID, hd.ID)
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
				raData.ID,
				*hd,
			)
			errorhandling.PrettifyErrorIfExists(err)

			// things to check:
			// 1. relayer folder exists
			isRelayerDirPresent, err := filesystem.DirNotEmpty(relayerHome)
			if err != nil {
				pterm.Error.Printf("failed to check %s: %v\n", relayerHome, err)
				return
			}

			var isRelayerIbcPathValid bool

			if isRelayerDirPresent {
				isRelayerIbcPathValid, err = relayerutils.ValidateIbcPathChains(
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

			if !isRelayerIbcPathValid {
				pterm.Warning.Println("relayer config verification failed...")
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

			err = relayerutils.EnsureKeysArePresentAndFunded(*rollappChainData)
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

			err = relayerutils.EnsureKeysArePresentAndFunded(*rollappChainData)
			if err != nil {
				pterm.Error.Println("failed to ensure relayer keys are created/funded:", err)
				return
			}

			pterm.Info.Println("let's create that IBC connection, shall we?")
			seq := sequencer.GetInstance(*rollappChainData)

			dymintutils.WaitForHealthyRollApp("http://localhost:26657/health")
			err = relayer.WaitForValidRollappHeight(seq)
			if err != nil {
				pterm.Error.Printf("rollapp did not reach valid height: %v\n", err)
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

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}
