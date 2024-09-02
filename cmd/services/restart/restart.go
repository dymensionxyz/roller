package restart

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func Cmd(services []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart",
		Short: "Restarts the systemd services relevant to RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			err := restartSystemdServices(services)
			if err != nil {
				pterm.Error.Println("failed to restart systemd services:", err)
				return
			}
		},
	}
	return cmd
}

func restartSystemdServices(services []string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf(
			"the services commands are only available on linux machines",
		)
	}
	for _, service := range services {
		err := servicemanager.RestartSystemdService(fmt.Sprintf("%s.service", service))
		if err != nil {
			return fmt.Errorf("failed to restart %s systemd service: %v", service, err)
		}
	}
	pterm.Success.Printf(
		"💈 Services %s restarted successfully.\n",
		strings.Join(services, ", "),
	)
	return nil
}
