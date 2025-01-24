package oracle

import (
	"github.com/dymensionxyz/roller/cmd/oracle/priceoracle"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "Commands related to RollApp's component observability",
	}

	cmd.AddCommand(priceoracle.Cmd())

	return cmd
}
