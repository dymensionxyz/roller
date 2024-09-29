package keys

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/keys/export"
	importkeys "github.com/dymensionxyz/roller/cmd/rollapp/keys/import"
	"github.com/dymensionxyz/roller/cmd/rollapp/keys/list"
	"github.com/dymensionxyz/roller/cmd/rollapp/keys/showunarmoredprivkey"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Commands for managing the roller different keys.",
	}
	cmd.AddCommand(list.Cmd())
	cmd.AddCommand(showunarmoredprivkey.Cmd())
	cmd.AddCommand(export.Cmd())
	cmd.AddCommand(importkeys.Cmd())

	return cmd
}
