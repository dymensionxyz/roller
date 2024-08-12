package start

import (
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func RollappCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "load",
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
			for _, service := range []string{"rollapp", "da"} {
				err := servicemanager.StartSystemdService(service)
				if err != nil {
					pterm.Error.Printf("failed to start %s systemd service: %v", service, err)
					return
				}
			}
			pterm.Success.Println(
				"💈 Services 'sequencer', 'da-light-client' started successfully.",
			)
		},
	}
	return cmd
}
