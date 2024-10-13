package start

import (
	"fmt"
	"runtime"
	"strings"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/migrations"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/upgrades"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func RollappCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the systemd services on local machine",
		Run: func(cmd *cobra.Command, args []string) {
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollappConfig, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			if rollappConfig.HubData.ID != consts.MockHubID {
				raUpgrade, err := upgrades.NewRollappUpgrade(string(rollappConfig.RollappVMType))
				if err != nil {
					pterm.Error.Println("failed to check rollapp version equality: ", err)
				}

				err = migrations.RequireRollappMigrateIfNeeded(
					raUpgrade.CurrentVersionCommit,
					rollappConfig.RollappBinaryVersion,
					string(rollappConfig.RollappVMType),
				)
				if err != nil {
					pterm.Error.Println(err)
					return
				}
			}

			if runtime.GOOS == "darwin" {
				err := startLaunchctlServices(consts.RollappSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start launchd services:", err)
					return
				}
			} else if runtime.GOOS == "linux" {
				err := startSystemdServices(consts.RollappSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start systemd services:", err)
					return
				}
			} else {
				pterm.Info.Printf(
					"the %s commands currently support only darwin and linux operating systems",
					cmd.Use,
				)
				return
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s to set up IBC channels and start relaying packets\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller relayer setup"),
				)
				pterm.Info.Printf(
					"run %s to view the logs  of the rollapp\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller rollapp services logs"),
				)
			}()
		},
	}
	return cmd
}

func RelayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Starts the relayer locally",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS == "linux" {
				err := startSystemdServices(consts.RelayerSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start systemd services:", err)
					return
				}
			} else if runtime.GOOS == "darwin" {
				err := startLaunchctlServices(consts.RelayerSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start launchd services:", err)
					return
				}
			} else {
				pterm.Error.Printf(
					"the %s commands currently support only darwin and linux operating systems",
					cmd.Use,
				)
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s to join the eibc market\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller eibc init"),
				)
				pterm.Info.Printf(
					"run %s to view the current status of the relayer\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller relayer services logs"),
				)
			}()
		},
	}
	return cmd
}

func EibcCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the systemd services on local machine",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS == "linux" {
				err := startSystemdServices(consts.EibcSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start systemd services:", err)
					return
				}
			} else if runtime.GOOS == "darwin" {
				err := startLaunchctlServices(consts.EibcSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start launchd services:", err)
					return
				}
			} else {
				pterm.Error.Printf(
					"the %s commands currently support only darwin and linux operating systems",
					cmd.Use,
				)
			}

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Println(
					"that's all folks",
				)

				if runtime.GOOS == "linux" {
					pterm.Info.Printf(
						"run %s to view the current status of the eibc client\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("journalctl -fu eibc"),
					)
				}
			}()

		},
	}
	return cmd
}

func startSystemdServices(services []string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf(
			"the services commands are only available on linux machines",
		)
	}
	for _, service := range services {
		err := servicemanager.StartSystemdService(
			fmt.Sprintf("%s.service", service),
		)
		if err != nil {
			return fmt.Errorf("failed to start %s systemd service: %v", service, err)
		}
	}
	pterm.Success.Printf(
		"💈 Services %s started successfully.\n",
		strings.Join(services, ", "),
	)
	return nil
}

func startLaunchctlServices(services []string) error {
	for _, service := range services {
		err := servicemanager.StartLaunchctlService(
			service,
		)
		if err != nil {
			return fmt.Errorf("failed to start %s launchctl service: %v", service, err)
		}
	}
	pterm.Success.Printf(
		"💈 Services %s started successfully.\n",
		strings.Join(services, ", "),
	)
	return nil
}
