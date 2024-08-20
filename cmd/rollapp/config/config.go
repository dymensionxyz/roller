package config

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/config/set"
	"github.com/dymensionxyz/roller/cmd/rollapp/config/show"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Update the relevant configuration values related to RollApp",
	}

	cmd.AddCommand(show.Cmd())
	cmd.AddCommand(set.Cmd())

	return cmd
}
