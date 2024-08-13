package services

import "github.com/spf13/cobra"

func Cmd(loadCmd, startCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Commands for managing the rollapp services.",
	}
	cmd.AddCommand(loadCmd)
	cmd.AddCommand(startCmd)
	return cmd
}
