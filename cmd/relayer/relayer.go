package relayer

import (
	"github.com/dymensionxyz/roller/cmd/relayer/run"
	"github.com/dymensionxyz/roller/cmd/relayer/start"
	"github.com/dymensionxyz/roller/cmd/relayer/status"
	"github.com/dymensionxyz/roller/cmd/services"
	loadservices "github.com/dymensionxyz/roller/cmd/services/load"
	restartservices "github.com/dymensionxyz/roller/cmd/services/restart"
	startservices "github.com/dymensionxyz/roller/cmd/services/start"
	stopservices "github.com/dymensionxyz/roller/cmd/services/stop"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relayer",
		Short: "Commands for running and managing the RollApp relayer.",
	}

	cmd.AddCommand(run.Cmd())
	cmd.AddCommand(start.Cmd())
	cmd.AddCommand(status.Cmd())

	sl := []string{"relayer"}
	cmd.AddCommand(
		services.Cmd(
			loadservices.Cmd(sl, cmd.Use),
			startservices.RelayerCmd(),
			restartservices.Cmd(sl),
			stopservices.Cmd(sl),
		),
	)

	return cmd
}
