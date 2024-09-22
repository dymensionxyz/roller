package start

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

// var OneDaySequencePrice = big.NewInt(1)

var (
	RollappDirPath string
	LogPath        string
	DaLcEndpoint   string
	DaLogPath      string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the RollApp node interactively",
		Long: `Start the RollApp node interactively.

Consider using 'services' if you want to run a 'systemd' service instead.
`,
		Run: func(cmd *cobra.Command, args []string) {
			showSequencerBalance, _ := cmd.Flags().GetBool("show-sequencer-balance")
			logLevel, _ := cmd.Flags().GetString("log-level")
			logLevels := []string{"debug", "info", "warn", "error", "fatal"}
			if !slices.Contains(logLevels, logLevel) {
				logLevel = "debug"
			}

			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}
			home, err := filesystem.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			seq := sequencer.GetInstance(rollappConfig)
			startRollappCmd := seq.GetStartCmd(logLevel)

			fmt.Println(startRollappCmd.String())

			LogPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
			RollappDirPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			go bash.RunCmdAsync(
				ctx,
				startRollappCmd,
				func() {
					PrintOutput(
						rollappConfig,
						strconv.Itoa(startRollappCmd.Process.Pid),
						showSequencerBalance,
						true,
						true,
						true,
					)
					err := createPidFile(RollappDirPath, startRollappCmd)
					if err != nil {
						pterm.Warning.Println("failed to create pid file")
					}
				},
				parseError,
				utils.WithLogging(utils.GetSequencerLogPath(rollappConfig)),
			)

			select {}
		},
	}
	cmd.Flags().Bool("show-sequencer-balance", false, "initialize the rollapp with mock backend")
	cmd.Flags().String("log-level", "debug", "pass the log level to the rollapp")

	return cmd
}

func PrintOutput(
	rlpCfg config.RollappConfig,
	pid string,
	withBalance,
	withEndpoints,
	withProcessInfo,
	isHealthy bool,
) {
	logPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
	rollappDirPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp)

	seq := sequencer.GetInstance(rlpCfg)

	var msg string
	if isHealthy {
		msg = pterm.DefaultBasicText.WithStyle(
			pterm.
				FgGreen.ToStyle(),
		).Sprintf("💈 The Rollapp %s is running on your local machine!", rlpCfg.NodeType)
	} else {
		msg = pterm.DefaultBasicText.WithStyle(
			pterm.
				FgRed.ToStyle(),
		).Sprintf(
			"❗ The Rollapp %s is in unhealthy state. Please check the logs for more information.",
			rlpCfg.NodeType,
		)
	}

	fmt.Println(msg)
	pterm.Println()
	fmt.Printf(
		"💈 RollApp ID: %s\n", pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprintf(rlpCfg.RollappID),
	)

	if withEndpoints {
		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Endpoints:")
		if rlpCfg.VMType == "evm" {
			fmt.Printf("EVM RPC: http://0.0.0.0:%v\n", seq.JsonRPCPort)
		}
		fmt.Printf("Node RPC: http://0.0.0.0:%v\n", seq.RPCPort)
		fmt.Printf("Rest API: http://0.0.0.0:%v\n", seq.APIPort)
	}

	pterm.DefaultSection.WithIndentCharacter("💈").
		Println("Filesystem Paths:")
	fmt.Println("Rollapp root dir: ", rollappDirPath)
	fmt.Println("Log file path: ", logPath)

	if withProcessInfo {
		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Process Info:")
		fmt.Println("PID: ", pid)
	}

	if isHealthy {
		seqAddrData, err := sequencerutils.GetSequencerData(rlpCfg)
		daManager := datalayer.NewDAManager(consts.Celestia, rlpCfg.Home)
		celAddrData, errCel := daManager.GetDAAccData(rlpCfg)
		if err != nil {
			return
		}

		if err != nil {
			return
		}
		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Wallet Info:")
		fmt.Println("Sequencer Address:", seqAddrData[0].Address)
		if withBalance && rlpCfg.NodeType == "sequencer" {
			fmt.Println("Sequencer Balance:", seqAddrData[0].Balance.String())
		}

		if errCel != nil {
			pterm.Error.Println("failed to retrieve DA address")
			return
		}

		fmt.Println("Da Address:", celAddrData[0].Address)
		if withBalance && rlpCfg.NodeType == "sequencer" && rlpCfg.HubData.ID != "mock" {
			fmt.Println("Da Balance:", celAddrData[0].Balance.String())
		}
	}
}

func printDaOutput() {
	fmt.Println("💈 The data availability light node is running on your local machine!")
	fmt.Printf("💈 Light node endpoint: %s\n", DaLcEndpoint)
	fmt.Printf("💈 Log file path: %s\n", DaLogPath)
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
