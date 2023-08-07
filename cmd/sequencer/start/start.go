package sequnecer_start

import (
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

// TODO: Test sequencing on 35-C and update the price
var OneDaySequencePrice = big.NewInt(1)

var (
	RollappDirPath string
	LogPath        string
)

func StartCmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "start",
		Short: "Runs the rollapp sequencer.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			LogPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
			RollappDirPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)
			sequencerInsufficientAddrs, err := utils.GetSequencerInsufficientAddrs(rollappConfig, OneDaySequencePrice)
			utils.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(sequencerInsufficientAddrs, rollappConfig)
			startRollappCmd := GetStartRollappCmd(rollappConfig)
			utils.RunBashCmdAsync(startRollappCmd, printOutput, parseError,
				utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)))
		},
	}

	return runCmd
}

func printOutput() {
	fmt.Println("💈 The Rollapp sequencer is running on your local machine!")
	fmt.Println("💈 Default endpoints:")

	fmt.Println("💈 EVM RPC: http://0.0.0.0:8545")
	fmt.Println("💈 Node RPC: http://0.0.0.0:26657")
	fmt.Println("💈 Rest API: http://0.0.0.0:1317")

	fmt.Println("💈 Log file path: ", LogPath)
	fmt.Println("💈 Rollapp root dir: ", RollappDirPath)
}

func parseError(errMsg string) string {
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 && lines[0] == "Error: failed to initialize database: resource temporarily unavailable" {
		return "The Rollapp sequencer is already running on your local machine. Only one sequencer can run at any given time."
	}
	return errMsg
}

func GetStartRollappCmd(rollappConfig config.RollappConfig) *exec.Cmd {
	rollappConfigDir := filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)
	cmd := exec.Command(
		rollappConfig.RollappBinary,
		"start",
		"--home", rollappConfigDir,
		"--log-file", filepath.Join(rollappConfigDir, "rollapp.log"),
	)
	return cmd
}
