package blockexplorer

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/block-explorer/run"
	"github.com/dymensionxyz/roller/cmd/block-explorer/teardown"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-explorer",
		Short: "Commands for managing block explorer.",
	}

	cmd.AddCommand(run.Cmd())
	cmd.AddCommand(teardown.Cmd())

	return cmd
}
