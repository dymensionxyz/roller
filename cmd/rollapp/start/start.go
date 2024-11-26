package start

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/healthagent"
	"github.com/dymensionxyz/roller/utils/logging"
	"github.com/dymensionxyz/roller/utils/migrations"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
	"github.com/dymensionxyz/roller/utils/upgrades"
)

// var OneDaySequencePrice = big.NewInt(1)

var (
	DaLcEndpoint string
	DaLogPath    string
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the RollApp node interactively",
		Long: `Start the RollApp node interactively.

Consider using 'services' if you want to run a 'systemd'(unix) or 'launchd'(mac) service instead.
`,
		Run: func(cmd *cobra.Command, args []string) {
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

			if rollappConfig.HubData.ID != consts.MockHubID {
				raUpgrade, err := upgrades.NewRollappUpgrade(string(rollappConfig.RollappVMType))
				if err != nil {
					pterm.Error.Println("failed to check rollapp version equality: ", err)
				}

				err = migrations.RequireRollappMigrateIfNeeded(
					raUpgrade.CurrentVersionCommit[:6],
					// rollappConfig.RollappBinaryVersion[:6],
					"main",
					string(rollappConfig.RollappVMType),
				)
				if err != nil {
					pterm.Info.Println(err)
					// return // TODO: enable it after handling proper versions in rollapp-evm fork
				}
			}

			seq := sequencer.GetInstance(rollappConfig)
			startRollappCmd := seq.GetStartCmd(logLevel, rollappConfig.KeyringBackend)

			fmt.Println(startRollappCmd.String())

			rollerLogger := logging.GetRollerLogger(rollappConfig.Home)

			if rollappConfig.HubData.ID != "mock" {
				go healthagent.Start(home, rollerLogger)
			}

			done := make(chan error, 1)
			// nolint: errcheck
			if rollappConfig.KeyringBackend == consts.SupportedKeyringBackends.OS {
				pswFileName, err := filesystem.GetOsKeyringPswFileName(
					consts.Executables.RollappEVM,
				)
				if err != nil {
					pterm.Error.Println("failed to get os keyring password file name: ", err)
					return
				}

				fp := filepath.Join(home, string(pswFileName))
				psw, err := filesystem.ReadFromFile(fp)
				if err != nil {
					pterm.Error.Println("failed to read os keyring password file: ", err)
					return
				}

				pr := map[string]string{
					"Enter keyring passphrase":    psw,
					"Re-enter keyring passphrase": psw,
				}

				ctx, cancel := context.WithCancel(cmd.Context())
				defer cancel()
				go func() {
					err := bash.ExecCmdFollow(
						done,
						ctx,
						startRollappCmd,
						pr,
					)

					done <- err
				}()
			} else {
				ctx, cancel := context.WithCancel(cmd.Context())
				defer cancel()

				go func() {
					err := bash.ExecCmdFollow(
						done,
						ctx,
						startRollappCmd,
						nil, // No need for printOutput since we configured output above
					)

					done <- err
				}()

				select {
				case err := <-done:
					if err != nil {
						pterm.Error.Println("rollapp's process returned an error: ", err)
						os.Exit(1)
					}
				case <-ctx.Done():
					pterm.Error.Println("context cancelled, terminating command")
					return
				}

			}

			select {}
		},
	}
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
	fmt.Println("log file path.......", logPath, rollappDirPath)

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
			Sprint(rlpCfg.RollappID),
	)
	fmt.Printf(
		"💈 Keyring Backend: %s\n", pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprint(rlpCfg.KeyringBackend),
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

	fmt.Println("is healthyyyyyyyy........", isHealthy)

	if isHealthy {
		seqAddrData, err := sequencerutils.GetSequencerData(rlpCfg)
		daManager := datalayer.NewDAManager(consts.Avail, rlpCfg.Home, rlpCfg.KeyringBackend) // avail as a da
		fmt.Println("da manager heree.........", daManager, rlpCfg.KeyringBackend, err)
		availAddrData, errCel := daManager.GetDAAccData(rlpCfg)
		fmt.Println("avail configggg........", rlpCfg.DA)
		if err != nil {
			return
		}

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
	// nolint: errcheck
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
