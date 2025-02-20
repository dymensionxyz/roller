package start

import (
	"runtime"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func AlertAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the alert-agent systemd service on local machine",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS == "linux" {
				err := startSystemdServices(consts.AlertAgentSystemdServices)
				if err != nil {
					pterm.Error.Println("failed to start systemd services:", err)
					return
				}
			} else if runtime.GOOS == "darwin" {
				err := startLaunchctlServices(consts.AlertAgentSystemdServices)
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
						"run %s to view the current status of the alert-agent\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("journalctl -fu alert-agent"),
					)
				}
			}()
		},
	}
	return cmd
}
