package order

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order",
		Short: "Commands related to fulfillment of eibc orders",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("orders related to fulfillment of eibc orders")

			// dymd tx eibc fulfill-order 41faa715d0e3484ed6e4a9d7bd9e4965fe1851951defdf9f6720e7686b477883 --from client --home ~/.fulfiller --keyring-backend test --node https://rpc-dymension.mzonder.com:443 --chain-id dymension_110
			// 0-1 -b block --fees 1000000000000000adym
		},
	}

	return cmd
}
