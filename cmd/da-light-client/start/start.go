package start

import (
	"context"
	"errors"
	"fmt"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

const (
	rpcEndpointFlag     = "rpc-endpoint"
	metricsEndpointFlag = "metrics-endpoint"
)

var LCEndpoint = ""

var LogFilePath = ""

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the DA light client.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			pterm.Info.Println("loading roller config file")
			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			// TODO: refactor the version comparison for migrations
			// errorhandling.RequireMigrateIfNeeded(rollappConfig)

			metricsEndpoint := cmd.Flag(metricsEndpointFlag).Value.String()
			if metricsEndpoint != "" && rollappConfig.DA.Backend != consts.Celestia {
				errorhandling.PrettifyErrorIfExists(
					errors.New("metrics endpoint can only be set for celestia"),
				)
			}
			damanager := datalayer.NewDAManager(rollappConfig.DA.Backend, rollappConfig.Home)

			pterm.Info.Println("checking for da address balance")
			insufficientBalances, err := damanager.CheckDABalance()
			errorhandling.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(insufficientBalances)

			damanager.SetRPCEndpoint(rollappConfig.DA.StateNode)
			if metricsEndpoint != "" {
				damanager.SetMetricsEndpoint(metricsEndpoint)
			}

			startDALCCmd := damanager.GetStartDACmd()
			if startDALCCmd == nil {
				errorhandling.PrettifyErrorIfExists(
					errors.New(
						"DA doesn't need to run separately. It runs automatically with the app",
					),
				)
			}

			LogFilePath = utils.GetDALogFilePath(rollappConfig.Home)
			LCEndpoint = damanager.GetLightNodeEndpoint()
			ctx, cancel := context.WithCancel(context.Background())

			fmt.Println(startDALCCmd.String())
			defer cancel()
			go bash.RunCmdAsync(
				ctx,
				startDALCCmd,
				printOutput,
				parseError,
				utils.WithLogging(LogFilePath),
			)
			select {}
		},
	}

	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().
		StringP(rpcEndpointFlag, "", consts.DefaultCelestiaStateNode, "The DA rpc endpoint to connect to.")
	cmd.Flags().
		StringP(metricsEndpointFlag, "", "", "The OTEL collector metrics endpoint to connect to.")
}

func printOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Printf("💈 Light node endpoint: %s\n", LCEndpoint)
	fmt.Printf("💈 Log file path: %s\n", LogFilePath)
}

func parseError(errMsg string) string {
	return errMsg
}
