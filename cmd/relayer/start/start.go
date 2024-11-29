package start

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/logging"
	"github.com/dymensionxyz/roller/utils/rollapp"
)

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

			var rlyCfg relayer.Config
			err := rlyCfg.Load(rlyConfigPath)
			if err != nil {
				pterm.Error.Println("failed to load relayer config: ", err)
				return
			}

			raData := rlyCfg.RaDataFromRelayerConfig()
			hd := rlyCfg.HubDataFromRelayerConfig()

			raResponse, err := rollapp.GetMetadataFromChain(raData.ID, *hd)
			if err != nil {
				pterm.Error.Println("failed to fetch rollapp information from hub: ", err)
				return
			}
			raData.Denom = raResponse.Rollapp.GenesisInfo.NativeDenom.Base

			err = relayer.VerifyRelayerBalances(*hd)
			if err != nil {
				pterm.Error.Println("failed to check balances", err)
				return
			}
			relayerLogFilePath := logging.GetRelayerLogPath(home)
			logger := logging.GetLogger(relayerLogFilePath)
			logFileOption := logging.WithLoggerLogging(logger)
			rly := relayer.NewRelayer(
				home,
				*raData,
				*hd,
			)
			rly.SetLogger(logger)

			err = rly.LoadActiveChannel(*raData, *hd)
			errorhandling.PrettifyErrorIfExists(err)

			// override := cmd.Flag(flagOverride).Changed
			//
			// if override {
			// 	fmt.Println("💈 Overriding the existing relayer channel")
			// }

			if rly.ChannelReady() {
				fmt.Println("💈 IBC transfer channel is established!")
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
				"💈 The relayer is running successfully on you local machine!\nChannels:\nRollApp: %s\n<->\nHub: %s\n",
				rly.DstChannel,
				rly.SrcChannel,
			)
			fmt.Println("💈 Log file path: ", relayerLogFilePath)

			select {}
		},
	}

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}
