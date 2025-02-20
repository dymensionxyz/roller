package start

import (
	"github.com/spf13/cobra"
)

func AlertAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Alert Agent systemd service",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement alert agent service start logic
			return nil
		},
	}

	return cmd
}
