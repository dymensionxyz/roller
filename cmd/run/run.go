package run

import (
	"context"
	"os"
	"os/exec"
	"sync"

	"github.com/dymensionxyz/roller/cmd/consts"
	relayer_start "github.com/dymensionxyz/roller/cmd/relayer/start"
	sequnecer_start "github.com/dymensionxyz/roller/cmd/sequencer/start"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the rollapp on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			verifyBalances(rollappConfig)
			logger := utils.GetRollerLogger(rollappConfig.Home)
			ctx, cancel := context.WithCancel(context.Background())
			waitingGroup := sync.WaitGroup{}
			waitingGroup.Add(3)
			serviceConfig := utils.ServiceConfig{
				Logger:    logger,
				Context:   ctx,
				WaitGroup: &waitingGroup,
			}

			/* ------------------------------ run processes ----------------------------- */
			runDaWithRestarts(rollappConfig, serviceConfig)
			runSequencerWithRestarts(rollappConfig, serviceConfig)
			runRelayerWithRestarts(rollappConfig, serviceConfig)

			/* ------------------------------ render output ----------------------------- */
			RenderUI(rollappConfig)
			cancel()
			waitingGroup.Wait()
		},
	}

	return cmd
}

func runRelayerWithRestarts(config config.RollappConfig, serviceConfig utils.ServiceConfig) {
	startRelayerCmd := getStartRelayerCmd(config)
	utils.RunServiceWithRestart(startRelayerCmd, serviceConfig)
}

func getStartRelayerCmd(config config.RollappConfig) *exec.Cmd {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return exec.Command(ex, "relayer", "start", "--home", config.Home)
}

func runDaWithRestarts(rollappConfig config.RollappConfig, serviceConfig utils.ServiceConfig) {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	daLogFilePath := utils.GetDALogFilePath(rollappConfig.Home)
	startDALCCmd := damanager.GetStartDACmd()
	if startDALCCmd == nil {
		serviceConfig.WaitGroup.Done()
		return
	}
	utils.RunServiceWithRestart(startDALCCmd, serviceConfig, utils.WithLogging(daLogFilePath))
}

func runSequencerWithRestarts(rollappConfig config.RollappConfig, serviceConfig utils.ServiceConfig) {
	startRollappCmd := sequnecer_start.GetStartRollappCmd(rollappConfig, consts.DefaultDALCRPC)
	utils.RunServiceWithRestart(startRollappCmd, serviceConfig, utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)))
}

func verifyBalances(rollappConfig config.RollappConfig) {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	insufficientBalances, err := damanager.CheckDABalance()
	utils.PrettifyErrorIfExists(err)

	sequencerInsufficientBalances, err := utils.GetSequencerInsufficientAddrs(
		rollappConfig, *sequnecer_start.OneDaySequencePrice)
	utils.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, sequencerInsufficientBalances...)

	rlyAddrs, err := relayer_start.GetRlyHubInsufficientBalances(rollappConfig)
	utils.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, rlyAddrs...)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances, rollappConfig)
}
