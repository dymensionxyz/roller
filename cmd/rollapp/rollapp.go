package rollapp

import (
	"github.com/spf13/cobra"

	initrollapp "github.com/dymensionxyz/roller/cmd/rollapp/init"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rollapp [command]",
		Short: "Commands to initialize and run a RollApp",
	}

	cmd.AddCommand(initrollapp.Cmd)

	return cmd
}
