package rollapp

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/config"
	initrollapp "github.com/dymensionxyz/roller/cmd/rollapp/init"
	"github.com/dymensionxyz/roller/cmd/rollapp/run"
	"github.com/dymensionxyz/roller/cmd/rollapp/start"
	"github.com/dymensionxyz/roller/cmd/rollapp/status"
	"github.com/dymensionxyz/roller/cmd/services"
	"github.com/dymensionxyz/roller/cmd/services/load"
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
	cmd.AddCommand(run.Cmd())
	cmd.AddCommand(services.Cmd(load.RollappCmd()))

	return cmd
}
