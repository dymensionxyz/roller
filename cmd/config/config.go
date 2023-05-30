package config

import (
	configInit "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/spf13/cobra"
)

func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for setting up and managing rollapp configuration files.",
	}
	cmd.AddCommand(configInit.InitCmd())
	return cmd
}
