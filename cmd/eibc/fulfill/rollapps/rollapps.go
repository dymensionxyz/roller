package rollapps

import (
	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/rollapps/list"
	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/rollapps/remove"
	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/rollapps/set"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollapps",
		Short: "Commands to manage the whitelist of RollApps to fulfill eibc orders for",
		Args:  cobra.MaximumNArgs(1),
	}

	cmd.AddCommand(set.Cmd())
	cmd.AddCommand(list.Cmd())
	cmd.AddCommand(remove.Cmd())

	return cmd
}
