package keys

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/keys/export"
	"github.com/dymensionxyz/roller/cmd/keys/list"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Commands for managing the roller different keys.",
	}
	cmd.AddCommand(list.Cmd())
	cmd.AddCommand(export.Cmd())
	return cmd
}
