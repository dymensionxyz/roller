package config

import (
	configInit "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/config/set"
	"github.com/dymensionxyz/roller/cmd/config/show"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for setting up and managing rollapp configuration files.",
	}
	cmd.AddCommand(configInit.InitCmd())
	cmd.AddCommand(show.Cmd())
	cmd.AddCommand(set.Cmd())
	return cmd
}
