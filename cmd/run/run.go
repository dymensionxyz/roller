package run

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"os/exec"
	"path/filepath"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

const daLightClientEndpointFlag = "da-light-client-endpoint"

func RunCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := initconfig.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			startRollappCmd := getStartRollapCmd(rollappConfig, cmd.Flag(daLightClientEndpointFlag).Value.String())
			startRollappErr := startRollappCmd.Run()
			utils.PrettifyErrorIfExists(startRollappErr)
		},
	}
	addFlags(runCmd)
	utils.AddGlobalFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(daLightClientEndpointFlag, "", "http://localhost:26659", "The DA light client endpoint.")
}

func getStartRollapCmd(rollappConfig initconfig.InitConfig, daLightClientEndpoint string) *exec.Cmd {
	daConfig := fmt.Sprintf(`{"base_url": "%s", "timeout": 60000000000, "fee":20000, "gas_limit": 20000000, "namespace_id":[0,0,0,0,0,0,255,255]}`, daLightClientEndpoint)
	rollappConfigDir := filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)
	settlementConfig := fmt.Sprintf(`{"node_address": "%s", "rollapp_id": "%s", "dym_account_name": "%s", "keyring_home_dir": "%s", "keyring_backend":"test", "gas_fees": "2000000udym"}`,
		rollappConfig.HubData.RPC_URL, rollappConfig.RollappID, consts.KeyNames.HubSequencer, rollappConfigDir)

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
