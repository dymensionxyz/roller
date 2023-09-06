package sequnecer_start

import (
	"fmt"
	"github.com/dymensionxyz/roller/sequencer"
	"math/big"
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
			utils.RequireMigrateIfNeeded(rollappConfig)
			LogPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
			RollappDirPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)
			sequencerInsufficientAddrs, err := utils.GetSequencerInsufficientAddrs(rollappConfig, OneDaySequencePrice)
			utils.PrettifyErrorIfExists(err)
			utils.PrintInsufficientBalancesIfAny(sequencerInsufficientAddrs, rollappConfig)
			seq := sequencer.GetInstance(rollappConfig)
			startRollappCmd := seq.GetStartCmd()
			utils.RunBashCmdAsync(startRollappCmd, func() {
				printOutput(rollappConfig)
			}, parseError,
				utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)))
		},
	}

	return runCmd
}

func printOutput(rlpCfg config.RollappConfig) {
	seq := sequencer.GetInstance(rlpCfg)
	fmt.Println("💈 The Rollapp sequencer is running on your local machine!")
	fmt.Println("💈 Endpoints:")

	fmt.Printf("💈 EVM RPC: http://0.0.0.0:%v\n", seq.JsonRPCPort)
	fmt.Printf("💈 Node RPC: http://0.0.0.0:%v\n", seq.RPCPort)
	fmt.Printf("💈 Rest API: http://0.0.0.0:%v\n", seq.APIPort)

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
