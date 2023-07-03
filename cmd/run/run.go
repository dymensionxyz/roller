package run

import (
	"context"
	"os"
	"os/exec"
	"sync"

	"github.com/dymensionxyz/roller/cmd/consts"
	da_start "github.com/dymensionxyz/roller/cmd/da-light-client/start"
	relayer_start "github.com/dymensionxyz/roller/cmd/relayer/start"
	sequnecer_start "github.com/dymensionxyz/roller/cmd/sequencer/start"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the rollapp on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
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
			runDaWithRestarts(rollappConfig, serviceConfig)
			runSequencerWithRestarts(rollappConfig, serviceConfig)
			runRelayerWithRestarts(rollappConfig, serviceConfig)
			PrintServicesStatus(rollappConfig)
			cancel()
			logger.Println("Killing subprocesses")
			waitingGroup.Wait()
			logger.Println("killed them")
		},
	}

	return cmd
}

func runRelayerWithRestarts(config utils.RollappConfig, serviceConfig utils.ServiceConfig) {
	startRelayerCmd := getStartRelayerCmd(config)
	utils.RunServiceWithRestart(startRelayerCmd, serviceConfig)
}

func getStartRelayerCmd(config utils.RollappConfig) *exec.Cmd {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return exec.Command(ex, "relayer", "start", "--home", config.Home)
}

func runDaWithRestarts(rollappConfig utils.RollappConfig, serviceConfig utils.ServiceConfig) {
	daLogFilePath := utils.GetDALogFilePath(rollappConfig.Home)
	startDALCCmd := da_start.GetStartDACmd(rollappConfig, consts.DefaultCelestiaRPC)
	utils.RunServiceWithRestart(startDALCCmd, serviceConfig, utils.WithLogging(daLogFilePath))
}

func runSequencerWithRestarts(rollappConfig utils.RollappConfig, serviceConfig utils.ServiceConfig) {
	startRollappCmd := sequnecer_start.GetStartRollappCmd(rollappConfig, consts.DefaultDALCRPC)
	utils.RunServiceWithRestart(startRollappCmd, serviceConfig, utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)))
}

func verifyBalances(rollappConfig utils.RollappConfig) {
	insufficientBalances, err := da_start.CheckDABalance(rollappConfig)
	utils.PrettifyErrorIfExists(err)
	sequencerInsufficientBalances, err := utils.GetSequencerInsufficientAddrs(
		rollappConfig, *sequnecer_start.OneDaySequencePrice)
	utils.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, sequencerInsufficientBalances...)
	rlyAddrs, err := relayer_start.GetRlyHubInsufficientBalances(rollappConfig)
	utils.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, rlyAddrs...)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)
}
