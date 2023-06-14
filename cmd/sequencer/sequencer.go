package sequencer

import (
	start "github.com/dymensionxyz/roller/cmd/sequencer/start"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sequencer",
		Short: "Commands for running and managing the RollApp sequnecer.",
	}
	cmd.AddCommand(start.StartCmd())
	return cmd
}
