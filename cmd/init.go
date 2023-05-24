package cmd

import (
	"context"
	"fmt"

	// Import NewClientFromNode from client utils.go
	cosmos_client "github.com/cosmos/cosmos-sdk/client"

	"github.com/dymensionxyz/dymension/x/rollapp/types"
	"github.com/spf13/cobra"
)

const hubRPC = "https://rpc-hub-35c.dymension.xyz:443"

// Create a function to validate a chain id string. it should return an error if the chain id is invalid.

func validateChainID(chainID string, cmd *cobra.Command) error {
	// create a new node flag already initiated with value http://localhost:26659
	cmd.Flags().String("node", "https://rpc-hub-35c.dymension.xyz:443", "")
	clientCtx := cosmos_client.GetClientContextFromCmd(cmd)
	rpcURI := "https://rpc-hub-35c.dymension.xyz:443"
	clientCtx = clientCtx.WithNodeURI(rpcURI)
	client, err := cosmos_client.NewClientFromNode(rpcURI)
	if err != nil {
		return err
	}

	clientCtx = clientCtx.WithClient(client)
	queryClient := types.NewQueryClient(clientCtx)
	// create a list of cmd.Flags() plus a "node" string flag I want to set here

	pageReq, _ := client.ReadPageRequest(cmd.Flags())
	params := &types.QueryAllRollappRequest{
		Pagination: pageReq,
	}
	res, err := queryClient.RollappAll(context.Background(), params)
	if err != nil {
		// print the error nicely
		fmt.Println("Error while querying the server:")
		fmt.Println(err)
	}
	// print the response nicely
	fmt.Println("Response from the server:")
	fmt.Println(res)
	println(fmt.Sprintf("Initializing a rollapp configuration for chain %s", chainID))
	return nil
}

var initCmd = &cobra.Command{
	Use:   "init <chain-id>",
	Short: "Initialize a rollapp configuration on your local machine",
	Run: func(cmd *cobra.Command, args []string) {
		validateChainID(args[0], cmd)
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
