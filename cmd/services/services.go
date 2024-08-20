package services

import "github.com/spf13/cobra"

func Cmd(loadCmd, startCmd, restartCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services [command]",
		Short: "Commands for managing systemd services.",
	}
	cmd.AddCommand(loadCmd)
	cmd.AddCommand(startCmd)
	cmd.AddCommand(restartCmd)
	return cmd
}
