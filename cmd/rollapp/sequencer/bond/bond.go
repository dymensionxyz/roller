package bond

import (
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond/get"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond/set"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond/unbond"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond",
		Short: "Commands to manage sequencer bond",
	}

	cmd.AddCommand(get.Cmd())
	cmd.AddCommand(set.Cmd())
	cmd.AddCommand(unbond.Cmd())

	return cmd
}
