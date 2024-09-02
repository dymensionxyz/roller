package fulfill

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/order"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fulfill",
		Short: "Commands related to fulfillment of eibc orders",
	}

	cmd.AddCommand(order.Cmd())

	return cmd
}
