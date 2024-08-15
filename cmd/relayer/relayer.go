package relayer

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/relayer/run"
	"github.com/dymensionxyz/roller/cmd/relayer/start"
	"github.com/dymensionxyz/roller/cmd/relayer/status"
	"github.com/dymensionxyz/roller/cmd/services"
	loadservices "github.com/dymensionxyz/roller/cmd/services/load"
	startservices "github.com/dymensionxyz/roller/cmd/services/start"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relayer",
		Short: "Commands for running and managing the RollApp relayer.",
	}
	cmd.AddCommand(run.Cmd())
	cmd.AddCommand(start.Cmd())
	cmd.AddCommand(status.Cmd())
	cmd.AddCommand(services.Cmd(loadservices.RelayerCmd(), startservices.RelayerCmd()))

	return cmd
}
