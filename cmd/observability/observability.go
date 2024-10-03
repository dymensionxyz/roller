package observability

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/observability/export"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observability",
		Short: "Commands related to RollApp's component observability",
		Run: func(cmd *cobra.Command, args []string) {
		},
	}

	cmd.AddCommand(export.Cmd())

	return cmd
}
