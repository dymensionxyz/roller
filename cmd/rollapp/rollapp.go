package rollapp

import (
	"github.com/dymensionxyz/roller/cmd/rollapp/config"
	initrollapp "github.com/dymensionxyz/roller/cmd/rollapp/init"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer"
	"github.com/dymensionxyz/roller/cmd/rollapp/setup"
	"github.com/dymensionxyz/roller/cmd/rollapp/start"
	"github.com/dymensionxyz/roller/cmd/rollapp/status"
	"github.com/dymensionxyz/roller/cmd/services"
	loadservices "github.com/dymensionxyz/roller/cmd/services/load"
	restartservices "github.com/dymensionxyz/roller/cmd/services/restart"
	startservices "github.com/dymensionxyz/roller/cmd/services/start"
	stopservices "github.com/dymensionxyz/roller/cmd/services/stop"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollapp [command]",
		Short: "Commands to initialize and run a RollApp",
	}

	cmd.AddCommand(initrollapp.Cmd())
	cmd.AddCommand(status.Cmd())
	cmd.AddCommand(start.Cmd())
	cmd.AddCommand(config.Cmd())
	cmd.AddCommand(setup.Cmd())
	cmd.AddCommand(sequencer.Cmd())

	sl := []string{"rollapp", "da-light-client"}
	cmd.AddCommand(
		services.Cmd(
			loadservices.Cmd(sl, "rollapp"),
			startservices.RollappCmd(),
			restartservices.Cmd(sl),
			stopservices.Cmd(sl),
		),
	)

	return cmd
}
