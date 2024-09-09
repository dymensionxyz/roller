package sequencer

import (
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/metadata"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/rewards"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sequencer [command]",
		Short: "Commands to manage sequencer instance",
	}

	cmd.AddCommand(metadata.Cmd())
	cmd.AddCommand(rewards.Cmd())
	cmd.AddCommand(bond.Cmd())

	return cmd
}
