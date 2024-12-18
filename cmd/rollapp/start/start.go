package start

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/healthagent"
	"github.com/dymensionxyz/roller/utils/logging"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
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

			// if rollappConfig.HubData.ID != consts.MockHubID {
			// 	raUpgrade, err := upgrades.NewRollappUpgrade(string(rollappConfig.RollappVMType))
			// 	if err != nil {
			// 		pterm.Error.Println("failed to check rollapp version equality: ", err)
			// 	}

			// 	err = migrations.RequireRollappMigrateIfNeeded(
			// 		raUpgrade.CurrentVersionCommit[:6],
			// 		rollappConfig.RollappBinaryVersion[:6],
			// 		string(rollappConfig.RollappVMType),
			// 	)
			// 	if err != nil {
			// 		pterm.Error.Println(err)
			// 		return
			// 	}
			// }

			err = sequencerutils.CheckBalance(rollappConfig)
			if err != nil {
				pterm.Error.Println("failed to check sequencer balance: ", err)
				return
			}

			seq := sequencer.GetInstance(rollappConfig)
			startRollappCmd := seq.GetStartCmd(logLevel, rollappConfig.KeyringBackend)
			fmt.Println(startRollappCmd.String())

			rollerLogger := logging.GetRollerLogger(rollappConfig.Home)

			if rollappConfig.HubData.ID != "mock" && rollappConfig.HealthAgent.Enabled {
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
