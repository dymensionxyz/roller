package config

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/config/set"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Commands for setting up and managing rollapp configuration files.",
	}

	// cmd.AddCommand(show.Cmd())
	cmd.AddCommand(set.Cmd())
	// cmd.AddCommand(export.Cmd())
	return cmd
}
