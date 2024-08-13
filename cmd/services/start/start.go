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
		Short: "Loads the different rollapp services on the local machine",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "linux" {
				pterm.Error.Printf(
					"the %s commands are only available on linux machines",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("'services'"),
				)

				return
			}
			services := []string{"rollapp", "da-light-client"}
			for _, service := range services {
				err := servicemanager.StartSystemdService(fmt.Sprintf("%s.service", service))
				if err != nil {
					pterm.Error.Printf("failed to start %s systemd service: %v\n", service, err)
					return
				}
			}
			pterm.Success.Printf(
				"💈 Services %s started successfully.\n",
				strings.Join(services, ", "),
			)
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
