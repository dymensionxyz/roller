package start

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
	"math/big"
	"os/exec"
	"path/filepath"
)

const rpcEndpointFlag = "--rpc-endpoint"

// TODO: test how much is enough to run the LC for one day and set the minimum balance accordingly.
var lcMinBalance = big.NewInt(1)

const gatewayAddr = "0.0.0.0"
const gatewayPort = "26659"
const celestiaRestApiEndpoint = "https://api-arabica-8.consensus.celestia-arabica.com"

var LCEndpoint = fmt.Sprintf("http://%s:%s", gatewayAddr, gatewayPort)

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			verifyDABalanceRest(rollappConfig)
			rpcEndpoint := cmd.Flag(rpcEndpointFlag).Value.String()
			startDACmd := getStartCelestiaLCCmd(rollappConfig, rpcEndpoint)
			logFilePath := filepath.Join(rollappConfig.Home, consts.ConfigDirName.DALightNode, "light_client.log")
			utils.RunBashCmdAsync(startDACmd, printOutput, parseError, utils.WithLogging(logFilePath))
		},
	}
	utils.AddGlobalFlags(runCmd)
	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(rpcEndpointFlag, "", "consensus-full-arabica-8.celestia-arabica.com",
		"The DA rpc endpoint to connect to.")
}

func verifyDABalanceRest(config utils.RollappConfig) {
	celAddress, err := utils.GetCelestiaAddress(config.Home)
	utils.PrettifyErrorIfExists(err)
	var restQueryUrl = fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		celestiaRestApiEndpoint, celAddress,
	)
	balancesJson, err := utils.RestQueryJson(restQueryUrl)
	utils.PrettifyErrorIfExists(err)
	balance, err := utils.ParseBalanceFromResponse(*balancesJson, consts.Denoms.Celestia)
	utils.PrettifyErrorIfExists(err)
	if balance.Cmp(lcMinBalance) < 0 {
		outputInsufficientBalanceError(balance, celAddress)
	}
}

func outputInsufficientBalanceError(currentBalance *big.Int, celAddress string) {
	insufficientBalances := make([]utils.NotFundedAddressData, 0)
	insufficientBalances = append(insufficientBalances, utils.NotFundedAddressData{
		Address:         celAddress,
		CurrentBalance:  currentBalance,
		RequiredBalance: lcMinBalance,
		KeyName:         consts.KeyNames.DALightNode,
		Denom:           consts.Denoms.Celestia,
	})
	utils.PrintInsufficientBalancesIfAny(insufficientBalances)
}

func printOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Printf("💈 Light node endpoint: %s", LCEndpoint)
}

func parseError(errMsg string) string {
	return errMsg
}

func getStartCelestiaLCCmd(rollappConfig utils.RollappConfig, rpcEndpoint string) *exec.Cmd {
	return exec.Command(
		consts.Executables.Celestia, "light", "start",
		"--core.ip", rpcEndpoint,
		"--node.store", filepath.Join(rollappConfig.Home, consts.ConfigDirName.DALightNode),
		"--gateway",
		"--gateway.addr", gatewayAddr,
		"--gateway.port", gatewayPort,
		"--p2p.network", "arabica",
	)
}
