package run

import (
	"context"
	"os"
	"os/exec"
	"sync"

	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"

	relayer_start "github.com/dymensionxyz/roller/cmd/relayer/start"
	sequnecer_start "github.com/dymensionxyz/roller/cmd/sequencer/start"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
	"github.com/spf13/cobra"
)

var flagNames = struct {
	NoOutput string
}{
	NoOutput: "no-output",
}

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
			seq := sequencer.GetInstance(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			runSequencerWithRestarts(seq, serviceConfig)
			runDaWithRestarts(rollappConfig, serviceConfig)
			runRelayerWithRestarts(rollappConfig, serviceConfig)

			/* ------------------------------ render output ----------------------------- */
			noOutput, err := cmd.Flags().GetBool(flagNames.NoOutput)
			utils.PrettifyErrorIfExists(err)
			if noOutput {
				select {}
			} else {
				RenderUI(rollappConfig, serviceConfig)
			}
			cancel()
			spin := utils.GetLoadingSpinner()
			spin.Suffix = " Stopping rollapp services, please wait..."
			spin.Start()
			utils.RunOnInterrupt(spin.Stop)
			waitingGroup.Wait()
			spin.Stop()
		},
	}
	cmd.Flags().BoolP(flagNames.NoOutput, "", false, "Run the rollapp without rendering the UI.")
	return cmd
}

func runRelayerWithRestarts(cfg config.RollappConfig, serviceConfig *servicemanager.ServiceConfig) {
	startRelayerCmd := getStartRelayerCmd(cfg)

	rly := relayer.NewRelayer(cfg.Home, cfg.RollappID, cfg.HubData.ID)

	service := servicemanager.Service{
		Command:  startRelayerCmd,
		FetchFn:  relayer.GetRelayerAccountsData,
		UIData:   servicemanager.UIData{Name: "Relayer"},
		StatusFn: rly.GetRelayerStatus,
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
	damanager.SetRPCEndpoint(celestia.DefaultCelestiaRPC)
	daLogFilePath := utils.GetDALogFilePath(rollappConfig.Home)
	service := servicemanager.Service{
		Command:  damanager.GetStartDACmd(),
		FetchFn:  damanager.GetDAAccData,
		StatusFn: damanager.GetStatus,
		UIData:   servicemanager.UIData{Name: "DA Light Client"},
	}
	serviceConfig.AddService("DA Light Client", service)
	serviceConfig.RunServiceWithRestart("DA Light Client", utils.WithLogging(daLogFilePath))
}

func runSequencerWithRestarts(seq *sequencer.Sequencer, serviceConfig *servicemanager.ServiceConfig) {
	startRollappCmd := seq.GetStartCmd()

	service := servicemanager.Service{
		Command:  startRollappCmd,
		FetchFn:  utils.GetSequencerData,
		StatusFn: seq.GetSequencerStatus,
		UIData:   servicemanager.UIData{Name: "Sequencer"},
	}
	serviceConfig.AddService("Sequencer", service)
	serviceConfig.RunServiceWithRestart("Sequencer", utils.WithLogging(utils.GetSequencerLogPath(seq.RlpCfg)))
}

func verifyBalances(rollappConfig config.RollappConfig) {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	insufficientBalances, err := damanager.CheckDABalance()
	utils.PrettifyErrorIfExists(err)

	sequencerInsufficientBalances, err := utils.GetSequencerInsufficientAddrs(
		rollappConfig, sequnecer_start.OneDaySequencePrice)
	utils.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, sequencerInsufficientBalances...)

	rlyAddrs, err := relayer_start.GetRlyHubInsufficientBalances(rollappConfig)
	utils.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, rlyAddrs...)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances, rollappConfig)
}
