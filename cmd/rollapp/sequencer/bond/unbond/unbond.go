package unbond

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "Commands to manage sequencer instance",
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("This command is not implemented yet.")
		},
	}

	return cmd
}
