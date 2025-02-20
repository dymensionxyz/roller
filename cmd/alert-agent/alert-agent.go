package alertagent

import (
	"github.com/spf13/cobra"

	initam "github.com/dymensionxyz/roller/cmd/alert-agent/init"
	"github.com/dymensionxyz/roller/cmd/alert-agent/start"
	"github.com/dymensionxyz/roller/cmd/services"
	"github.com/dymensionxyz/roller/cmd/services/load"
	"github.com/dymensionxyz/roller/cmd/services/restart"
	startservices "github.com/dymensionxyz/roller/cmd/services/start"
	"github.com/dymensionxyz/roller/cmd/services/stop"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alert-agent",
		Short: "Commands related to Alert Agent operations",
	}

	cmd.AddCommand(start.Cmd())
	cmd.AddCommand(initam.Cmd())

	sl := []string{"alert-agent"}
	cmd.AddCommand(
		services.Cmd(
			load.Cmd(sl, "alert-agent"),
			startservices.AlertAgentCmd(),
			restart.Cmd(sl),
			stop.Cmd(sl),
		),
	)

	return cmd
}
