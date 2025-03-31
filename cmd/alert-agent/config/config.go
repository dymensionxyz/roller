package config

import (
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Update the relevant configuration values related to Alert Agent",
	}

	cmd.AddCommand(AddEibcLPCmd())
	cmd.AddCommand(RemoveEibcLPCmd())
	return cmd
}
