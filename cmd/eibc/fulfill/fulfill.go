package fulfill

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/order"
	"github.com/dymensionxyz/roller/cmd/eibc/fulfill/rollapps"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "fulfill",
		Short: "Commands related to fulfillment of eibc orders",
	}

	cmd.AddCommand(order.Cmd())
	cmd.AddCommand(rollapps.Cmd())

	return cmd
}
