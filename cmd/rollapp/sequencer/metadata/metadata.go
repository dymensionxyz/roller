package metadata

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/metadata/export"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/metadata/update"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metadata [command]",
		Short: "Commands to manage sequencer metadata",
	}

	cmd.AddCommand(export.Cmd())
	cmd.AddCommand(update.Cmd())

	return cmd
}
