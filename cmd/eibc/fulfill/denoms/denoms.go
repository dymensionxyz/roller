package denoms

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/denoms/list"
	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/denoms/remove"
	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/denoms/set"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "denoms",
		Short: "Commands to manage the whitelist of Denoms to fulfill eibc orders for",
		Args:  cobra.MaximumNArgs(1),
	}

	cmd.AddCommand(list.Cmd())
	cmd.AddCommand(remove.Cmd())
	cmd.AddCommand(set.Cmd())

	return cmd
}
