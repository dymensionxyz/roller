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
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
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
			logger := utils.GetRollerLogger(rollappConfig.Home)

			ctx, cancel := context.WithCancel(context.Background())
			waitingGroup := sync.WaitGroup{}
			serviceConfig := &servicemanager.ServiceConfig{
				Logger:    logger,
				Context:   ctx,
				WaitGroup: &waitingGroup,
			}
			/* ----------------------------- verify balances ---------------------------- */
			verifyBalances(rollappConfig)

			/* ------------------------------ run processes ----------------------------- */
			runDaWithRestarts(rollappConfig, serviceConfig)
			runSequencerWithRestarts(rollappConfig, serviceConfig)
			runRelayerWithRestarts(rollappConfig, serviceConfig)

			/* ------------------------------ render output ----------------------------- */
			RenderUI(rollappConfig, serviceConfig)
			cancel()
			waitingGroup.Wait()
		},
	}

	return cmd
}

func runRelayerWithRestarts(cfg config.RollappConfig, serviceConfig *servicemanager.ServiceConfig) {
	startRelayerCmd := getStartRelayerCmd(cfg)
	service := servicemanager.ServiceData{
		Command: startRelayerCmd,
		FetchFn: utils.GetRelayerAddresses,
		UIData:  servicemanager.UIData{Name: "Relayer"},
	}
	serviceConfig.AddService("Relayer", service)
	serviceConfig.RunServiceWithRestart("Relayer")
}

func getStartRelayerCmd(config config.RollappConfig) *exec.Cmd {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return exec.Command(ex, "relayer", "start", "--home", config.Home)
}

func runDaWithRestarts(rollappConfig config.RollappConfig, serviceConfig *servicemanager.ServiceConfig) {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	daLogFilePath := utils.GetDALogFilePath(rollappConfig.Home)
	startDALCCmd := damanager.GetStartDACmd(consts.DefaultCelestiaRPC)
	if startDALCCmd == nil {
		return
	}

	service := servicemanager.ServiceData{
		Command: startDALCCmd,
		FetchFn: damanager.GetDAAccData,
		UIData:  servicemanager.UIData{Name: "DA Light Client"},
	}
	serviceConfig.AddService("DA Light Client", service)
	serviceConfig.RunServiceWithRestart("DA Light Client", utils.WithLogging(daLogFilePath))
}

func runSequencerWithRestarts(rollappConfig config.RollappConfig, serviceConfig *servicemanager.ServiceConfig) {
	startRollappCmd := sequnecer_start.GetStartRollappCmd(rollappConfig, consts.DefaultDALCRPC)
	service := servicemanager.ServiceData{
		Command: startRollappCmd,
		FetchFn: utils.GetSequencerData,
		UIData:  servicemanager.UIData{Name: "Sequencer"},
	}
	serviceConfig.AddService("Sequencer", service)
	serviceConfig.RunServiceWithRestart("Sequencer", utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)))
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

	utils.PrintInsufficientBalancesIfAny(insufficientBalances)
}
