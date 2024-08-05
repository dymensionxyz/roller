package relayer

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/relayer/run"
	"github.com/dymensionxyz/roller/cmd/relayer/status"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relayer",
		Short: "Commands for running and managing the RollApp relayer.",
	}
	cmd.AddCommand(run.Cmd())
	// cmd.AddCommand(start.Cmd())
	cmd.AddCommand(status.Cmd())
	return cmd
}
