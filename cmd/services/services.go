package services

import "github.com/spf13/cobra"

// TODO: use options instead
func Cmd(loadCmd, startCmd, restartCmd, stopCmd, logsCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services [command]",
		Short: "Commands for managing systemd services.",
	}
	cmd.AddCommand(loadCmd)
	cmd.AddCommand(startCmd)
	cmd.AddCommand(restartCmd)
	cmd.AddCommand(stopCmd)
	cmd.AddCommand(logsCmd)
	return cmd
}
