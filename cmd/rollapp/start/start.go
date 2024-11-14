package start

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/healthagent"
	"github.com/dymensionxyz/roller/utils/logging"
	"github.com/dymensionxyz/roller/utils/roller"
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
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollappConfig, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config: ", err)
				return
			}

			if rollappConfig.HubData.ID != consts.MockHubID { //TODO : enable it if required
				// raUpgrade, err := upgrades.NewRollappUpgrade(string(rollappConfig.RollappVMType))
				// if err != nil {
				// 	pterm.Error.Println("failed to check rollapp version equality: ", err)
				// }

				// err = migrations.RequireRollappMigrateIfNeeded(
				// 	raUpgrade.CurrentVersionCommit,
				// 	rollappConfig.RollappBinaryVersion,
				// 	string(rollappConfig.RollappVMType),
				// )
				// if err != nil {
				// 	pterm.Error.Println(err)
				// 	return
				// }
			}

			seq := sequencer.GetInstance(rollappConfig)
			startRollappCmd := seq.GetStartCmd(logLevel)

			fmt.Println(startRollappCmd.String())

			LogPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp, "rollapp.log")
			RollappDirPath = filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp)

			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			rollerLogger := logging.GetRollerLogger(rollappConfig.Home)

			nodeID, err := dymint.GetNodeID(home)
			if err != nil {
				fmt.Println("failed to retrieve dymint node id:", err)
				return
			}

			if rollappConfig.HubData.ID != "mock" {
				go healthagent.Start(home, rollerLogger)
			}

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
						nodeID,
					)
					err := createPidFile(RollappDirPath, startRollappCmd)
					if err != nil {
						pterm.Warning.Println("failed to create pid file")
					}
				},
				parseError,
				logging.WithLogging(logging.GetSequencerLogPath(rollappConfig)),
			)

			select {}
		},
	}
	cmd.Flags().Bool("show-sequencer-balance", false, "initialize the rollapp with mock backend")
	cmd.Flags().String("log-level", "debug", "pass the log level to the rollapp")

	return cmd
}

func PrintOutput(
	rlpCfg roller.RollappConfig,
	pid string,
	withBalance,
	withEndpoints,
	withProcessInfo,
	isHealthy bool,
	dymintNodeID string,
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

	fmt.Printf(
		"💈 Node ID: %s\n", pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprint(dymintNodeID),
	)

	if withEndpoints {
		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Endpoints:")
		if rlpCfg.RollappVMType == "evm" {
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
		fmt.Println("PID:", pid)
		fmt.Println("OS:", runtime.GOOS)
		fmt.Println("Architecture:", runtime.GOARCH)
	}

	if isHealthy {
		seqAddrData, err := sequencerutils.GetSequencerData(rlpCfg)
		if err != nil {
			return
		}
		// daManager := datalayer.NewDAManager(consts.Celestia, rlpCfg.Home)
		daManager := datalayer.NewDAManager(consts.Avail, rlpCfg.Home)
		// fmt.Println("da manager.....", daManager, rlpCfg, rlpCfg.Home)
		availAddrData, errCel := daManager.GetDAAccData(rlpCfg)
		fmt.Println("avail address:: and error", availAddrData, errCel)
		if errCel != nil {
			pterm.Error.Println("failed to retrieve DA address") // here check
			return
		}

		pterm.DefaultSection.WithIndentCharacter("💈").
			Println("Wallet Info:")
		fmt.Println("Sequencer Address:", seqAddrData[0].Address, seqAddrData[0].Balance.String())
		if withBalance && rlpCfg.NodeType == "sequencer" {
			fmt.Println("Sequencer Balance:", seqAddrData[0].Balance.String())
		}

		// fmt.Println("Da Address:", availAddrData[0].Address)
		if withBalance && rlpCfg.NodeType == "sequencer" && rlpCfg.HubData.ID != "mock" {
			fmt.Println("Da Balance:", availAddrData[0].Balance.String())
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
