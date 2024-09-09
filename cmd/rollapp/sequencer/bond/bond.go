package bond

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond/decrease"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond/get"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond/increase"
	"github.com/dymensionxyz/roller/cmd/rollapp/sequencer/bond/unbond"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bond",
		Short: "Commands to manage sequencer bond",
	}

	cmd.AddCommand(get.Cmd())
	cmd.AddCommand(increase.Cmd())
	cmd.AddCommand(decrease.Cmd())
	cmd.AddCommand(unbond.Cmd())

	return cmd
}
