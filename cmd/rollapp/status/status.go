package status

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				fmt.Println("failed to load config:", err)
				return
			}
			errorhandling.PrettifyErrorIfExists(err)
			seq := sequencer.GetInstance(rollappConfig)
			fmt.Println(seq.GetSequencerStatus(rollappConfig))
		},
	}
	return cmd
}
