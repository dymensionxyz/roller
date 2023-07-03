package start

import (
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

// TODO: test how much is enough to run the LC for one day and set the minimum balance accordingly.
const (
	gatewayAddr     = "0.0.0.0"
	gatewayPort     = "26659"
	rpcEndpointFlag = "--rpc-endpoint"
)

var (
	lcMinBalance = big.NewInt(1)
	LCEndpoint   = fmt.Sprintf("http://%s:%s", gatewayAddr, gatewayPort)
)

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			insufficientBalances, err := CheckDABalance(rollappConfig)
			utils.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(insufficientBalances)
			rpcEndpoint := cmd.Flag(rpcEndpointFlag).Value.String()
			startDALCCmd := GetStartDACmd(rollappConfig, rpcEndpoint)
			logFilePath := utils.GetDALogFilePath(rollappConfig.Home)
			utils.RunBashCmdAsync(startDALCCmd, printOutput, parseError, utils.WithLogging(logFilePath))
		},
	}

	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(rpcEndpointFlag, "", consts.DefaultCelestiaRPC, "The DA rpc endpoint to connect to.")
}

func CheckDABalance(config utils.RollappConfig) ([]utils.NotFundedAddressData, error) {
	accData, err := utils.GetCelLCAccData(config)
	if err != nil {
		return nil, err
	}
	var insufficientBalances []utils.NotFundedAddressData
	if accData.Balance.Cmp(lcMinBalance) < 0 {
		insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
			Address:         accData.Address,
			CurrentBalance:  accData.Balance,
			RequiredBalance: lcMinBalance,
			KeyName:         consts.KeysIds.DALightNode,
			Denom:           consts.Denoms.Celestia,
		})
	}
	return insufficientBalances, nil
}

func printOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Printf("💈 Light node endpoint: %s", LCEndpoint)
}

func parseError(errMsg string) string {
	return errMsg
}

func GetStartDACmd(rollappConfig utils.RollappConfig, rpcEndpoint string) *exec.Cmd {
	return exec.Command(
		consts.Executables.Celestia, "light", "start",
		"--core.ip", rpcEndpoint,
		"--node.store", filepath.Join(rollappConfig.Home, consts.ConfigDirName.DALightNode),
		"--gateway",
		"--gateway.addr", gatewayAddr,
		"--gateway.port", gatewayPort,
		"--p2p.network", consts.DefaultCelestiaNetwork,
	)
}
