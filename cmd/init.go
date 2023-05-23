package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// define a string constant for for the hub rpc url
const hubRPC = "https://rpc-hub-35c.dymension.xyz:443"

var initCmd = &cobra.Command{
	Use:   "init <chain-id>",
	Short: "Initialize a rollapp configuration on your local machine",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("init called")
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	configCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("hub-rpc", "", hubRPC, "Dymension Hub rpc endpoint")
	initCmd.Flags().StringP("light-node-endpoint", "", "localhost:26659", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided.")
	initCmd.Flags().StringP("denom", "", "", "The rollapp token smallest denominator, for example `wei` in Ethereum.")
	initCmd.Flags().StringP("key-prefix", "", "", "The `bech32` prefix of the rollapp keys.")
	initCmd.Flags().StringP("rollapp-binary", "", "", "The rollapp binary. Should be passed only if you built a custom rollapp.")
	initCmd.Flags().Int64P("decimals", "", 18, "The number of decimal places a rollapp token supports.")
}
