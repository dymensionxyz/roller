package sequencer

import (
	sequnecer_start "github.com/dymensionxyz/roller/cmd/sequencer/start"
	"github.com/spf13/cobra"
)

func SequencerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sequencer",
		Short: "Commands for running and managing the RollApp sequnecer.",
	}
	cmd.AddCommand(sequnecer_start.StartCmd())
	return cmd
}
