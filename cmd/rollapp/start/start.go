package start

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

// var OneDaySequencePrice = big.NewInt(1)

var (
	RollappDirPath string
	LogPath        string
)

var LCEndpoint = ""

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the RollApp node interactively",
		Long: `Start the RollApp node interactively.

Consider using 'services' if you want to run a 'systemd' service instead.
`,
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}
			home, err := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			seq := sequencer.GetInstance(rollappConfig)
			startRollappCmd := seq.GetStartCmd()

			LogPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
			RollappDirPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)

			if runtime.GOOS != "linux" {
				damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
				startDALCCmd := damanager.GetStartDACmd()
				if startDALCCmd == nil {
					errorhandling.PrettifyErrorIfExists(
						errors.New(
							"DA doesn't need to run separately. It runs automatically with the app",
						),
					)
				}

				logFilePath := utils.GetDALogFilePath(rollappConfig.Home)
				LCEndpoint = damanager.GetLightNodeEndpoint()
				dactx, cancel := context.WithCancel(context.Background())

				fmt.Println(startDALCCmd.String())
				defer cancel()
				go bash.RunCmdAsync(
					dactx,
					startDALCCmd,
					printDALightClientOutput,
					parseError,
					utils.WithLogging(logFilePath),
				)
			}

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go bash.RunCmdAsync(
				ctx, startRollappCmd, func() {
					printOutput(rollappConfig, startRollappCmd)
					err := createPidFile(RollappDirPath, startRollappCmd)
					if err != nil {
						pterm.Warning.Println("failed to create pid file")
					}
				}, parseError,
				utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)),
			)

			select {}
		},
	}
	return cmd
}

func printOutput(rlpCfg config.RollappConfig, cmd *exec.Cmd) {
	seq := sequencer.GetInstance(rlpCfg)
	fmt.Println("💈 The Rollapp sequencer is running on your local machine!")
	fmt.Printf(
		"💈 RollApp ID: %s\n", pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprintf(rlpCfg.RollappID),
	)
	fmt.Println("💈 Endpoints:")

	fmt.Printf("💈 EVM RPC: http://0.0.0.0:%v\n", seq.JsonRPCPort)
	fmt.Printf("💈 Node RPC: http://0.0.0.0:%v\n", seq.RPCPort)
	fmt.Printf("💈 Rest API: http://0.0.0.0:%v\n", seq.APIPort)

	fmt.Println("💈 Log file path: ", LogPath)
	fmt.Println("💈 Rollapp root dir: ", RollappDirPath)
	fmt.Println("💈 PID: ", cmd.Process.Pid)
}

func createPidFile(path string, cmd *exec.Cmd) error {
	pidPath := filepath.Join(path, "rollapp.pid")
	file, err := os.Create(pidPath)
	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	// nolint errcheck
	defer file.Close()

	pid := cmd.Process.Pid
	_, err = file.WriteString(fmt.Sprintf("%d", pid))
	if err != nil {
		fmt.Println("Error writing to file:", err)
		return err
	}

	return nil
}

func parseError(errMsg string) string {
	lines := strings.Split(errMsg, "\n")
	if len(lines) > 0 &&
		lines[0] == "Error: failed to initialize database: resource temporarily unavailable" {
		return "The Rollapp sequencer is already running on your local machine. Only one sequencer can run at any given time."
	}
	return errMsg
}

func printDALightClientOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Printf("💈 Light node endpoint: %s", LCEndpoint)
	fmt.Println("💈 Log file path: ", LogPath)
}
