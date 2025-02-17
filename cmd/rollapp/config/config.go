package config

import (
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Update the relevant configuration values related to RollApp",
	}

	cmd.AddCommand(ShowCmd())
	cmd.AddCommand(SetCmd())

	return cmd
}
