package oracle

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/oracle/setup"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "Commands related to RollApp's component observability",
	}

	cmd.AddCommand(setup.Cmd())

	return cmd
}
