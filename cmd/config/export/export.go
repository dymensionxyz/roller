package export

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export the rollapp configurations jsons needed to list your rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			fmt.Println(rlpCfg)
		},
	}
	return cmd
}
