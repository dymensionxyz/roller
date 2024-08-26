package run

import (
	"context"
	"os"
	"os/exec"
	"sync"

	"github.com/spf13/cobra"

	relayerrun "github.com/dymensionxyz/roller/cmd/relayer/run"
	rollapprun "github.com/dymensionxyz/roller/cmd/rollapp/run"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
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
			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			errorhandling.RequireMigrateIfNeeded(rollappConfig)
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
			errorhandling.PrettifyErrorIfExists(err)
			runSequencerWithRestarts(seq, serviceConfig)
			runDaWithRestarts(rollappConfig, serviceConfig)
			runRelayerWithRestarts(rollappConfig, serviceConfig)

			/* ------------------------------ render output ----------------------------- */
			noOutput, err := cmd.Flags().GetBool(flagNames.NoOutput)
			errorhandling.PrettifyErrorIfExists(err)
			if noOutput {
				select {}
			} else {
				RenderUI(rollappConfig, serviceConfig)
			}
			cancel()
			spin := utils.GetLoadingSpinner()
			spin.Suffix = " Stopping rollapp services, please wait..."
			spin.Start()
			errorhandling.RunOnInterrupt(spin.Stop)
			waitingGroup.Wait()
			spin.Stop()
		},
	}
	cmd.Flags().BoolP(flagNames.NoOutput, "", false, "Run the rollapp without rendering the UI.")
	return cmd
}

func runRelayerWithRestarts(
	cfg config.RollappConfig,
	serviceConfig *servicemanager.ServiceConfig,
) {
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

func runDaWithRestarts(
	rollappConfig config.RollappConfig,
	serviceConfig *servicemanager.ServiceConfig,
) {
	damanager := datalayer.NewDAManager(rollappConfig.DA.Backend, rollappConfig.Home)
	damanager.SetRPCEndpoint(rollappConfig.DA.RpcUrl)
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

func runSequencerWithRestarts(
	seq *sequencer.Sequencer,
	serviceConfig *servicemanager.ServiceConfig,
) {
	startRollappCmd := seq.GetStartCmd()

	service := servicemanager.Service{
		Command:  startRollappCmd,
		FetchFn:  utils.GetSequencerData,
		StatusFn: seq.GetSequencerStatus,
		UIData:   servicemanager.UIData{Name: "Sequencer"},
	}
	serviceConfig.AddService("Sequencer", service)
	serviceConfig.RunServiceWithRestart(
		"Sequencer",
		utils.WithLogging(utils.GetSequencerLogPath(seq.RlpCfg)),
	)
}

func verifyBalances(rollappConfig config.RollappConfig) {
	damanager := datalayer.NewDAManager(rollappConfig.DA.Backend, rollappConfig.Home)
	insufficientBalances, err := damanager.CheckDABalance()
	errorhandling.PrettifyErrorIfExists(err)

	sequencerInsufficientBalances, err := utils.GetSequencerInsufficientAddrs(
		rollappConfig, rollapprun.OneDaySequencePrice,
	)
	errorhandling.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, sequencerInsufficientBalances...)

	rlyAddrs, err := relayerrun.GetRlyHubInsufficientBalances(rollappConfig)
	errorhandling.PrettifyErrorIfExists(err)
	insufficientBalances = append(insufficientBalances, rlyAddrs...)
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)
}
