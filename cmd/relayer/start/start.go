package start

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
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
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			// errorhandling.RequireMigrateIfNeeded(rollappConfig)

			VerifyRelayerBalances(rollappConfig)
			relayerLogFilePath := utils.GetRelayerLogPath(rollappConfig)
			logger := utils.GetLogger(relayerLogFilePath)
			logFileOption := utils.WithLoggerLogging(logger)
			rly := relayer.NewRelayer(
				rollappConfig.Home,
				rollappConfig.RollappID,
				rollappConfig.HubData.ID,
			)
			rly.SetLogger(logger)

			_, _, err = rly.LoadActiveChannel()
			errorhandling.PrettifyErrorIfExists(err)

			// override := cmd.Flag(flagOverride).Changed
			//
			// if override {
			// 	fmt.Println("💈 Overriding the existing relayer channel")
			// }

			if rly.ChannelReady() {
				fmt.Println("💈 IBC transfer channel is already established!")
				status := fmt.Sprintf("Active src, %s <-> %s, dst", rly.SrcChannel, rly.DstChannel)
				err := rly.WriteRelayerStatus(status)
				errorhandling.PrettifyErrorIfExists(err)
			} else {
				pterm.Error.Println("💈 No channels found, ensure you've setup the relayer")
				// seq := sequencer.GetInstance(rollappConfig)
				// _, err := rly.CreateIBCChannel(override, logFileOption, seq)
				// errorhandling.PrettifyErrorIfExists(err)
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
				"💈 The relayer is running successfully on you local machine!\nChannels: src, %s <-> %s, dst",
				rly.SrcChannel,
				rly.DstChannel,
			)

			select {}
		},
	}

	relayerStartCmd.Flags().
		BoolP(flagOverride, "", false, "override the existing relayer clients and channels")
	return relayerStartCmd
}

func VerifyRelayerBalances(rolCfg config.RollappConfig) {
	insufficientBalances, err := relayer.GetRelayerInsufficientBalances(rolCfg)
	errorhandling.PrettifyErrorIfExists(err)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)
}
