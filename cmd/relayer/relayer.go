package relayer

import (
	"github.com/dymensionxyz/roller/cmd/relayer/start"
	"github.com/dymensionxyz/roller/cmd/relayer/status"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "relayer",
		Short: "Commands for running and managing the RollApp relayer.",
	}
	cmd.AddCommand(start.Start())
	cmd.AddCommand(status.Cmd())
	return cmd
}
