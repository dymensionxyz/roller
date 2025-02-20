package start

import (
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start the Alert Agent service",
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: Implement alert agent start logic
			return nil
		},
	}

	return cmd
}
