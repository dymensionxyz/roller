package metrics

import "github.com/spf13/cobra"
import "github.com/dymensionxyz/roller/cmd/metrics/start"

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Commands for managing the roller Prometheus metrics server.",
	}
	cmd.AddCommand(start.Cmd())
	return cmd
}
