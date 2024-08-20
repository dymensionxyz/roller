package show

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	cmdutils "github.com/dymensionxyz/roller/cmd/utils"
	configutils "github.com/dymensionxyz/roller/utils/config"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "show",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(cmdutils.FlagNames.Home).Value.String()

			err := configutils.ShowCurrentConfigurableValues(home)
			if err != nil {
				pterm.Error.Println("failed to retrieve configurable values: ", err)
				return
			}
		},
	}

	return cmd
}
