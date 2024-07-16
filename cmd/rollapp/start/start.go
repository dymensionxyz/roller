package start

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)

			seq := sequencer.GetInstance(rollappConfig)

			c := seq.GetStartCmd()
			fmt.Println(c.String())

			err = utils.ExecBashCmdFollow(c)
			if err != nil {
				fmt.Println(err)
				return
			}
		},
	}
	return cmd
}
