package services

import "github.com/spf13/cobra"

func Cmd(loadCmd, startCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services [command]",
		Short: "Commands for managing systemd services.",
	}
	cmd.AddCommand(loadCmd)
	cmd.AddCommand(startCmd)
	return cmd
}
