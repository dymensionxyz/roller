package start

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/toml"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

const (
	rpcEndpointFlag     = "rpc-endpoint"
	metricsEndpointFlag = "metrics-endpoint"
)

var LCEndpoint = ""

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the DA light client.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := toml.LoadRollerConfigFromTOML(home)
			errorhandling.PrettifyErrorIfExists(err)
			errorhandling.RequireMigrateIfNeeded(rollappConfig)
			metricsEndpoint := cmd.Flag(metricsEndpointFlag).Value.String()
			if metricsEndpoint != "" && rollappConfig.DA != consts.Celestia {
				errorhandling.PrettifyErrorIfExists(
					errors.New("metrics endpoint can only be set for celestia"),
				)
			}
			damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)

			insufficientBalances, err := damanager.CheckDABalance()
			errorhandling.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(insufficientBalances)

			rpcEndpoint := cmd.Flag(rpcEndpointFlag).Value.String()
			if rpcEndpoint != "" {
				damanager.SetRPCEndpoint(rpcEndpoint)
			}
			if metricsEndpoint != "" {
				damanager.SetMetricsEndpoint(metricsEndpoint)
			}
			startDALCCmd := damanager.GetStartDACmd()
			if startDALCCmd == nil {
				errorhandling.PrettifyErrorIfExists(
					errors.New(
						"DA doesn't need to run seperatly. It runs automatically with the app",
					),
				)
			}
			logFilePath := utils.GetDALogFilePath(rollappConfig.Home)
			LCEndpoint = damanager.GetLightNodeEndpoint()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go bash.RunCmdAsync(
				ctx,
				startDALCCmd,
				printOutput,
				parseError,
				utils.WithLogging(logFilePath),
			)
			select {}
		},
	}

	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().
		StringP(rpcEndpointFlag, "", celestia.DefaultCelestiaRPC, "The DA rpc endpoint to connect to.")
	cmd.Flags().
		StringP(metricsEndpointFlag, "", "", "The OTEL collector metrics endpoint to connect to.")
}

func printOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Printf("💈 Light node endpoint: %s", LCEndpoint)
}

func parseError(errMsg string) string {
	return errMsg
}
