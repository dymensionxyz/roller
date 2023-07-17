package services

import (
	"github.com/dymensionxyz/roller/cmd/services/load"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Commands for managing the rollapp services.",
	}
	cmd.AddCommand(load.Cmd())
	return cmd
}
