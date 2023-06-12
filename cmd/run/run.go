package run

import (
	"fmt"
	"os/exec"
	"path/filepath"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func RunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := initconfig.LoadConfigFromTOML(home)
			initconfig.OutputCleanError(err)
			startRollappCmd := getStartRollapCmd(rollappConfig)
			startRollappErr := startRollappCmd.Run()
			initconfig.OutputCleanError(startRollappErr)
		},
	}
	utils.AddGlobalFlags(runCmd)
	return runCmd
}

func getStartRollapCmd(rollappConfig initconfig.InitConfig) *exec.Cmd {
	daConfig := fmt.Sprintf(`{"base_url": "%s", "timeout": 60000000000, "fee":20000, "gas_limit": 20000000, "namespace_id":[0,0,0,0,0,0,255,255]}`, rollappConfig.LightNodeEndpoint)
	rollappConfigDir := filepath.Join(rollappConfig.Home, initconfig.ConfigDirName.Rollapp)
	settlementConfig := fmt.Sprintf(`{"node_address": "%s", "rollapp_id": "%s", "dym_account_name": "%s", "keyring_home_dir": "%s", "keyring_backend":"test", "gas_fees": "2000000udym"}`, rollappConfig.HubData.RPC_URL, rollappConfig.RollappID, initconfig.
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
