package sequnecer_start

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"strings"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func StartCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := initconfig.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			LightNodeEndpoint := cmd.Flag(FlagNames.DAEndpoint).Value.String()
			startRollappCmd := getStartRollapCmd(rollappConfig, LightNodeEndpoint)
			utils.RunBashCmdAsync(startRollappCmd, printOutput, parseError)
		},
	}
	utils.AddGlobalFlags(runCmd)
	runCmd.Flags().StringP(FlagNames.DAEndpoint, "", "http://localhost:26659", "The data availability light node endpoint.")
	return runCmd
}

var FlagNames = struct {
	DAEndpoint string
}{
	DAEndpoint: "da-endpoint",
}

func printOutput() {
	fmt.Println("💈 The Rollapp sequencer is running on your local machine!")
	fmt.Println("💈 EVM RPC: http://0.0.0.0:8545")
	fmt.Println("💈 Node RPC: http://0.0.0.0:26657")
	fmt.Println("💈 Rest API: http://0.0.0.0:1317")
}

func parseError(errMsg string) string {
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 && lines[0] == "Error: failed to initialize database: resource temporarily unavailable" {
		return "The Rollapp sequencer is already running. Only one sequencer can run on the machine at any given time."
	}
	return errMsg
}

func getStartRollapCmd(rollappConfig initconfig.InitConfig, lightNodeEndpoint string) *exec.Cmd {
	daConfig := fmt.Sprintf(`{"base_url": "%s", "timeout": 60000000000, "fee":20000, "gas_limit": 20000000, "namespace_id":[0,0,0,0,0,0,255,255]}`,
		lightNodeEndpoint)
	rollappConfigDir := filepath.Join(rollappConfig.Home, initconfig.ConfigDirName.Rollapp)

	// TODO: Update the gas_fees to 2000000udym before 35-c launch.
	settlementConfig := fmt.Sprintf(`{"node_address": "%s", "rollapp_id": "%s", "dym_account_name": "%s", "keyring_home_dir": "%s", "keyring_backend":"test", "gas_fees": "0udym"}`, rollappConfig.HubData.RPC_URL, rollappConfig.RollappID, initconfig.
		KeyNames.HubSequencer, rollappConfigDir)

	return exec.Command(
		rollappConfig.RollappBinary, "start",
		"--dymint.aggregator",
		"--json-rpc.enable",
		"--json-rpc.api", "eth,txpool,personal,net,debug,web3,miner",
		"--dymint.da_layer", "celestia",
		"--dymint.da_config", daConfig,
		"--dymint.settlement_layer", "dymension",
		"--dymint.settlement_config", settlementConfig,
		"--dymint.block_batch_size", "1200",
		"--dymint.namespace_id", "000000000000ffff",
		"--dymint.block_time", "0.2s",
		"--home", rollappConfigDir,
		"--log_level", "debug",
		"--log-file", filepath.Join(rollappConfigDir, "rollapp.log"),
		"--max-log-size", "2000",
		"--module-log-level-override", "",
	)
}
