package stop

import (
	"fmt"
	"runtime"
	"strings"

	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd(services []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Short: "Stop the systemd services relevant to RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			err := stopSystemdServices(services)
			if err != nil {
				pterm.Error.Println("failed to restart systemd services:", err)
				return
			}
		},
	}
	return cmd
}

func stopSystemdServices(services []string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf(
			"the services commands are only available on linux machines",
		)
	}
	for _, service := range services {
		err := servicemanager.RestartSystemdService(fmt.Sprintf("%s.service", service))
		if err != nil {
			return fmt.Errorf("failed to stop %s systemd service: %v", service, err)
		}
	}
	pterm.Success.Printf(
		"💈 Services %s stopped successfully.\n",
		strings.Join(services, ", "),
	)
	return nil
}
