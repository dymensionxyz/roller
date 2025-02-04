package oracle

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/oracle/priceoracle"
	rngoracle "github.com/dymensionxyz/roller/cmd/oracle/rng"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "Commands related to RollApp's component observability",
	}

	cmd.AddCommand(priceoracle.Cmd())
	cmd.AddCommand(rngoracle.Cmd())

	return cmd
}
