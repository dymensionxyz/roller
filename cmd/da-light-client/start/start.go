package start

import (
	"errors"
	"fmt"
	"github.com/dymensionxyz/roller/data_layer/celestia"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/spf13/cobra"
)

const (
	rpcEndpointFlag = "--rpc-endpoint"
)

var LCEndpoint = ""

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the DA light client.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)

			damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
			insufficientBalances, err := damanager.CheckDABalance()
			utils.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(insufficientBalances, rollappConfig)

			rpcEndpoint := cmd.Flag(rpcEndpointFlag).Value.String()
			if rpcEndpoint != "" {
				damanager.SetRPCEndpoint(rpcEndpoint)
			}
			startDALCCmd := damanager.GetStartDACmd()
			if startDALCCmd == nil {
				utils.PrettifyErrorIfExists(errors.New("can't run mock DA. It runs automatically with the app"))
			}
			logFilePath := utils.GetDALogFilePath(rollappConfig.Home)
			LCEndpoint = damanager.GetLightNodeEndpoint()
			utils.RunBashCmdAsync(startDALCCmd, printOutput, parseError, utils.WithLogging(logFilePath))
		},
	}

	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(rpcEndpointFlag, "", celestia.DefaultCelestiaRPC, "The DA rpc endpoint to connect to.")
}

func printOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Printf("💈 Light node endpoint: %s", LCEndpoint)
}

func parseError(errMsg string) string {
	return errMsg
}
