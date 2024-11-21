package drs

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/drs/update"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drs",
		Short: "Commands related to drs(Dymension RollApp Standard).",
	}

	cmd.AddCommand(update.Cmd())

	return cmd
}
