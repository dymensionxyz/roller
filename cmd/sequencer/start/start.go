package sequnecer_start

import (
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

// TODO: Test sequencing on 35-C and update the price
var OneDaySequencePrice = big.NewInt(1)

func StartCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			sequencerInsufficientAddrs, err := utils.GetSequencerInsufficientAddrs(rollappConfig, *OneDaySequencePrice)
			utils.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(sequencerInsufficientAddrs)
			LightNodeEndpoint := cmd.Flag(FlagNames.DAEndpoint).Value.String()
			startRollappCmd := GetStartRollappCmd(rollappConfig, LightNodeEndpoint)
			utils.RunBashCmdAsync(startRollappCmd, printOutput, parseError, utils.WithLogging(
				filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")))
		},
	}
	utils.AddGlobalFlags(runCmd)
	runCmd.Flags().StringP(FlagNames.DAEndpoint, "", consts.DefaultDALCRPC,
		"The data availability light node endpoint.")
	return runCmd
}

var FlagNames = struct {
	DAEndpoint string
}{
	DAEndpoint: "da-endpoint",
}

func printOutput() {
	fmt.Println("💈 The Rollapp sequencer is running on your local machine!")

	//TODO: either mark the ports as default, or read from configuration file

	fmt.Println("💈 EVM RPC: http://0.0.0.0:8545")
	fmt.Println("💈 Node RPC: http://0.0.0.0:26657")
	fmt.Println("💈 Rest API: http://0.0.0.0:1317")

	//TODO: print the log file path
}

func parseError(errMsg string) string {
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 && lines[0] == "Error: failed to initialize database: resource temporarily unavailable" {
		return "The Rollapp sequencer is already running on your local machine. Only one sequencer can run at any given time."
	}
	return errMsg
}

func GetStartRollappCmd(rollappConfig utils.RollappConfig, lightNodeEndpoint string) *exec.Cmd {
	daConfig := fmt.Sprintf(`{"base_url": "%s", "timeout": 60000000000, "fee":20000, "gas_limit": 20000000, "namespace_id":"000000000000ffff"}`,
		lightNodeEndpoint)
	rollappConfigDir := filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)
	hubKeysDir := filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys)

	cmd := exec.Command(
		rollappConfig.RollappBinary, "start",
		"--dymint.da_layer", "celestia",
		"--dymint.da_config", daConfig,
		"--dymint.settlement_layer", "dymension",
		"--dymint.block_batch_size", "500",
		"--dymint.namespace_id", "000000000000ffff",
		"--dymint.block_time", "0.2s",
		"--dymint.batch_submit_max_time", "100s",
		"--dymint.empty_blocks_max_time", "10s",
		"--dymint.settlement_config.rollapp_id", rollappConfig.RollappID,
		"--dymint.settlement_config.node_address", rollappConfig.HubData.RPC_URL,
		"--dymint.settlement_config.dym_account_name", consts.KeysIds.HubSequencer,
		"--dymint.settlement_config.keyring_home_dir", hubKeysDir,
		"--dymint.settlement_config.gas_prices", "0udym",
		"--home", rollappConfigDir,
		"--log-file", filepath.Join(rollappConfigDir, "rollapp.log"),
		"--log_level", "info",
		"--max-log-size", "2000",
	)

	fmt.Println(cmd.String())
	return cmd
}
