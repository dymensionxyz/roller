package sequencer

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/metadata"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sequencer [command]",
		Short: "Commands to manage sequencer instance",
	}

	cmd.AddCommand(metadata.Cmd())
	cmd.AddCommand(bond.Cmd())

	return cmd
}
