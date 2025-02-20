package drs

import (
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "drs",
		Short: "Commands related to drs(Dymension RollApp Standard).",
	}

	cmd.AddCommand(UpdateCmd())
	cmd.AddCommand(UpgradeCmd())

	return cmd
}
