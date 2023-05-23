package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// define a string constant for for the hub rpc url
const hubRPC = "https://rpc-hub-35c.dymension.xyz:443"
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initialize a rollapp configuration on your local machine",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init called")
	},
}

func init() {
	configCmd.AddCommand(initCmd)
	/*
	Add those flags github copilot:
	- `**hub-rpc` (testnet hub)** - hub rpc endpoint.
- **`light-node-endpoint`(generated, localhost:26659)** - The data availability light node endpoint. Runs an Arabica Celestia light node if not provided.
- `**denom**` (**u + first three letters of the chain ID**) - The rollapp token denominator, for example `wei` in Ethereum.
- `**key-prefix**` (**$denom**)- The `bech32` prefix of the rollapp keys.
- `**rollapp-binary` (rollapp_evm)** - The rollapp binary.
- `**decimals**` ($10^{18}$) - Will be used to calculate the default total supply and staking parameters.
	 */
	 initCmd.Flags().StringP("hub-rpc", "", hubRPC, "Dymension Hub rpc endpoint")
}
