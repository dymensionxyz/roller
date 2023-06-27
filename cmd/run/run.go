package run

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	da_start "github.com/dymensionxyz/roller/cmd/da-light-client/start"
	relayer_start "github.com/dymensionxyz/roller/cmd/relayer/start"
	sequnecer_start "github.com/dymensionxyz/roller/cmd/sequencer/start"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"os/exec"
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
			runDaWithRestarts(rollappConfig)
			runSequencerWithRestarts(rollappConfig)
			runRelayerWithRestarts(rollappConfig)
			select {}
		},
	}
	utils.AddGlobalFlags(cmd)
	return cmd
}

func runRelayerWithRestarts(config utils.RollappConfig) {
	startRelayerCmd := getStartRelayerCmd(config)
	utils.RunCommandWithRestart(startRelayerCmd)
}

func getStartRelayerCmd(config utils.RollappConfig) *exec.Cmd {
	return exec.Command(consts.Executables.Roller, "relayer", "start", "--home", config.Home)
}

func runDaWithRestarts(rollappConfig utils.RollappConfig) {
	daLogFilePath := da_start.GetDALogFilePath(rollappConfig.Home)
	startDALCCmd := da_start.GetStartDACmd(rollappConfig, consts.DefaultCelestiaRPC)
	utils.RunCommandWithRestart(startDALCCmd, utils.WithLogging(daLogFilePath))
}

func runSequencerWithRestarts(rollappConfig utils.RollappConfig) {
	startRollappCmd := sequnecer_start.GetStartRollappCmd(rollappConfig, consts.DefaultDALCRPC)
	utils.RunCommandWithRestart(startRollappCmd)
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
