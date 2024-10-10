package show

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "show",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			err := roller.ShowCurrentConfigurableValues(home)
			if err != nil {
				pterm.Error.Println("failed to retrieve configurable values: ", err)
				return
			}
		},
	}

	return cmd
}
