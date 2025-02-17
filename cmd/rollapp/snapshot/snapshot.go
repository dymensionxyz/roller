package snapshot

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/snapshot/create"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshot",
		Short: "Commands to manage snapshot",
	}

	cmd.AddCommand(create.Cmd())

	return cmd
}
