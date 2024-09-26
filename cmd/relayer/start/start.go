package start

import (
	"context"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
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
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollerData, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			// errorhandling.RequireMigrateIfNeeded(rollappConfig)
			raRpc, err := sequencerutils.GetRpcEndpointFromChain(
				rollerData.RollappID,
				rollerData.HubData,
			)
			if err != nil {
				return
			}

			raData := consts.RollappData{
				ID:     rollerData.RollappID,
				RpcUrl: raRpc,
			}

			VerifyRelayerBalances(rollerData)
			relayerLogFilePath := utils.GetRelayerLogPath(home)
			logger := utils.GetLogger(relayerLogFilePath)
			logFileOption := utils.WithLoggerLogging(logger)
			rly := relayer.NewRelayer(
				rollerData.Home,
				rollerData.RollappID,
				rollerData.HubData.ID,
			)
			rly.SetLogger(logger)

			// TODO: relayer is initialized with both chains at this point and it should be possible
			// to construct the hub data from relayer config
			_, _, err = rly.LoadActiveChannel(raData, rollerData.HubData)
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
					rly.SrcChannel, rly.DstChannel,
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

func VerifyRelayerBalances(rolCfg config.RollappConfig) {
	insufficientBalances, err := relayer.GetRelayerInsufficientBalances(rolCfg)
	errorhandling.PrettifyErrorIfExists(err)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)
}
