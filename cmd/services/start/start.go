package start

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func RollappCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Loads the different RollApp services on the local machine",
		Run: func(cmd *cobra.Command, args []string) {
			services := []string{"rollapp", "da-light-client"}
			err := startSystemdServices(services)
			if err != nil {
				pterm.Error.Println("failed to start systemd services:", err)
				return
			}

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s to set up IBC channels and start relaying packets\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller relayer run"),
			)
		},
	}
	return cmd
}

func RelayerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Loads the different RollApp services on the local machine",
		Run: func(cmd *cobra.Command, args []string) {
			services := []string{"relayer"}
			err := startSystemdServices(services)
			if err != nil {
				pterm.Error.Println("failed to start systemd services:", err)
				return
			}

			pterm.Info.Println("next steps:")
			pterm.Info.Printf(
				"run %s to join the eibc market\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf("roller eibc run"),
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
		err := servicemanager.StartSystemdService(fmt.Sprintf("%s.service", service))
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
