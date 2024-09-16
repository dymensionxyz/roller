package start

import (
	"fmt"
	"runtime"
	"strings"

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

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s to join the eibc market\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller eibc init"),
			)
			pterm.Info.Printf(
				"run %s to view the current status of the relayer\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("journalctl -fu relayer"),
			)
		},
	}
	return cmd
}

func EibcCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the systemd services on local machine",
		Run: func(cmd *cobra.Command, args []string) {
			err := startSystemdServices(consts.EibcSystemdServices)
			if err != nil {
				pterm.Error.Println("failed to start systemd services:", err)
				return
			}

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"that's all folks",
			)
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
