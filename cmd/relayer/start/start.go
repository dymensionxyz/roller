package start

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/logging"
	relayerutils "github.com/dymensionxyz/roller/utils/relayer"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

// TODO: Test relaying on 35-C and update the prices

const (
	flagOverride = "override"
)

func Cmd() *cobra.Command {
	relayerStartCmd := &cobra.Command{
		Use:   "start",
		Short: "Start the relayer process interactively.",
		Long: `Start the relayer process interactively.

Consider using 'services' if you want to run a 'systemd' service instead.
`,
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rlyConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.Relayer,
				"config",
				"config.yaml",
			)

			data, err := os.ReadFile(rlyConfigPath)
			if err != nil {
				fmt.Printf("Error reading YAML file: %v\n", err)
				return
			}

			var rlyConfig relayerutils.Config
			err = yaml.Unmarshal(data, &rlyConfig)
			if err != nil {
				fmt.Printf("Error unmarshaling YAML: %v\n", err)
				return
			}

			// src is Hub, dst is RollApp
			raChainID := rlyConfig.Paths.HubRollapp.Dst.ChainID
			hubChainID := rlyConfig.Paths.HubRollapp.Src.ChainID

			_, hd, found := roller.FindHubDataByID(consts.Hubs, hubChainID)
			if !found {
				pterm.Error.Println("Hub Data not found for ", hubChainID)
				runCustomHub, _ := pterm.DefaultInteractiveConfirm.WithDefaultText("would you like to provide custom hub?").
					WithDefaultValue(false).
					Show()

				if !runCustomHub {
					pterm.Error.Println("cancelled by user")
					return
				}

				chd, err := config.CreateCustomHubData()
				if err != nil {
					pterm.Error.Println("failed to create custom hub data:", err)
					return
				}
				hd = *chd
			}

			raResponse, err := rollapp.GetMetadataFromChain(raChainID, hd)
			if err != nil {
				pterm.Error.Println("failed to fetch rollapp information from hub: ", err)
				return
			}

			// errorhandling.RequireMigrateIfNeeded(rollappConfig)
			raRpc, err := sequencerutils.GetRpcEndpointFromChain(
				raChainID,
				hd,
			)
			if err != nil {
				return
			}
			raData := consts.RollappData{
				ID:     raChainID,
				RpcUrl: fmt.Sprintf("%s:%d", raRpc, 443),
				Denom:  raResponse.Rollapp.GenesisInfo.NativeDenom.Base,
			}

			err = relayerutils.VerifyRelayerBalances(hd)
			if err != nil {
				pterm.Error.Println("failed to check balances", err)
				return
			}
			relayerLogFilePath := logging.GetRelayerLogPath(home)
			logger := logging.GetLogger(relayerLogFilePath)
			logFileOption := logging.WithLoggerLogging(logger)
			rly := relayer.NewRelayer(
				home,
				raChainID,
				hubChainID,
			)
			rly.SetLogger(logger)

			// TODO: relayer is initialized with both chains at this point and it should be possible
			// to construct the hub data from relayer config
			_, _, err = rly.LoadActiveChannel(raData, hd)
			errorhandling.PrettifyErrorIfExists(err)

			// override := cmd.Flag(flagOverride).Changed
			//
			// if override {
			// 	fmt.Println("💈 Overriding the existing relayer channel")
			// }

			if rly.ChannelReady() {
				fmt.Println("💈 IBC transfer channel is already established!")
				status := fmt.Sprintf(
					"Active\nrollapp: %s\n<->\nhub: %s\n",
					rly.DstChannel, rly.SrcChannel,
				)
				err := rly.WriteRelayerStatus(status)
				errorhandling.PrettifyErrorIfExists(err)
			} else {
				pterm.Error.Println("💈 No channels found, ensure you've setup the relayer")
				return
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			go bash.RunCmdAsync(
				ctx,
				rly.GetStartCmd(),
				func() {},
				func(errMessage string) string { return errMessage },
				logFileOption,
			)

			fmt.Printf(
				"💈 The relayer is running successfully on you local machine!\nChannels:\nrollapp: %s\n<->\nhub: %s\n",
				rly.SrcChannel,
				rly.DstChannel,
			)
			fmt.Println("💈 Log file path: ", relayerLogFilePath)

			select {}
		},
	}

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}
