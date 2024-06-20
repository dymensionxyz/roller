package status

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			seq := sequencer.GetInstance(rollappConfig)
			fmt.Println(seq.GetSequencerStatus(rollappConfig))
		},
	}
	return cmd
}
