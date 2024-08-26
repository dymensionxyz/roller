package blockexplorer

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/block-explorer/run"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "block-explorer",
		Short: "Commands for managing block explorer.",
	}

	cmd.AddCommand(run.Cmd())

	return cmd
}
