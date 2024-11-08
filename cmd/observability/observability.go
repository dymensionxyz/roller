package observability

import (
	_ "embed"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/observability/export"
	"github.com/dymensionxyz/roller/cmd/observability/query"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "observability",
		Short: "Commands related to RollApp's component observability",
	}

	cmd.AddCommand(export.Cmd())
	cmd.AddCommand(query.Cmd())

	return cmd
}
