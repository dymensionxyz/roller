package config

import (
	"github.com/dymensionxyz/roller/cmd/rollapp/config/show"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/rollapp/config/set"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Update the relevant configuration values related to rollapp",
	}

	cmd.AddCommand(show.Cmd())
	cmd.AddCommand(set.Cmd())

	return cmd
}
