package start

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func RollappCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the systemd services on local machine",
		Run: func(cmd *cobra.Command, args []string) {
			err := startSystemdServices(consts.RollappSystemdServices)
			if err != nil {
				pterm.Error.Println("failed to start systemd services:", err)
				return
			}

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s to set up IBC channels and start relaying packets\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller relayer setup"),
			)
			pterm.Info.Printf(
				"run %s to view the logs  of the relayer\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller relayer services load"),
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
			err := startSystemdServices(consts.RelayerSystemdServices)
			if err != nil {
				pterm.Error.Println("failed to start systemd services:", err)
				return
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
			"--show-sequencer-balance",
			"false",
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
