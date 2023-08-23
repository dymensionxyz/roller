package sequencer

import (
	sequnecer_start "github.com/dymensionxyz/roller/cmd/sequencer/start"
	"github.com/dymensionxyz/roller/cmd/sequencer/status"
	"github.com/spf13/cobra"
)

func SequencerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sequencer",
		Short: "Commands for running and managing the RollApp sequnecer.",
	}
	cmd.AddCommand(sequnecer_start.StartCmd())
	cmd.AddCommand(status.Cmd())
	return cmd
}
