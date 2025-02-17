package stop

import (
	"fmt"
	"runtime"
	"slices"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func Cmd(services []string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop",
		Args:  cobra.MaximumNArgs(1),
		Short: "Stop the systemd services relevant to RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			var servicesToStop []string
			if len(args) != 0 {
				if !slices.Contains(services, args[0]) {
					pterm.Error.Printf(
						"invalid service name %s. Available services: %v\n",
						args[0],
						services,
					)
					return
				}
				servicesToStop = []string{args[0]}
			} else {
				servicesToStop = services
			}
			err := stopSystemdServices(servicesToStop)
			if err != nil {
				pterm.Error.Println("failed to stop systemd services:", err)
				return
			}
		},
	}
	return cmd
}

func stopSystemdServices(services []string) error {
	if runtime.GOOS == "linux" {
		for _, service := range services {
			err := servicemanager.StopSystemdService(service)
			if err != nil {
				return fmt.Errorf("failed to stop %s systemd service: %v", service, err)
			}
		}
	} else if runtime.GOOS == "darwin" {
		for _, service := range services {
			err := servicemanager.StopLaunchdService(service)
			if err != nil {
				return fmt.Errorf("failed to stop %s systemd service: %v", service, err)
			}
		}
	} else {
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
	pterm.Success.Printf(
		"💈 Services %s stopped successfully.\n",
		strings.Join(services, ", "),
	)
	return nil
}
