package init

import (
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <chain-id>",
		Short: "Initialize a rollapp configuration on your local machine",
		Run: func(cmd *cobra.Command, args []string) {
			chainId := args[0]
			denom := args[1]
			rollappBinaryPath := getRollappBinaryPath(cmd.Flag(flagNames.RollappBinary).Value.String())
			decimals, err := cmd.Flags().GetUint64(flagNames.Decimals)
			if err != nil {
				panic(err)
			}
			generateKeys(!cmd.Flags().Changed(lightNodeEndpointFlag), chainId)
			if !cmd.Flags().Changed(lightNodeEndpointFlag) {
				if err = initializeLightNodeConfig(); err != nil {
					panic(err)
				}
			}
			initializeRollappConfig(rollappBinaryPath, chainId, denom)
			if err = initializeRollappGenesis(rollappBinaryPath, decimals, denom); err != nil {
				panic(err)
			}
		},
		Args: cobra.ExactArgs(2),
	}
	cmd.Flags().StringP(flagNames.HubRPC, "", hubRPC, "Dymension Hub rpc endpoint")
	cmd.Flags().StringP(flagNames.LightNodeEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided.")
	cmd.Flags().StringP("key-prefix", "", "", "The `bech32` prefix of the rollapp keys.")
	cmd.Flags().StringP("rollapp-binary", "", "", "The rollapp binary. Should be passed only if you built a custom rollapp.")
	cmd.Flags().Uint64P(flagNames.Decimals, "", 18, "The number of decimal places a rollapp token supports.")
	return cmd
}

func getRollappBinaryPath(rollappBinaryPath string) string {
	if rollappBinaryPath == "" {
		rollappBinaryPath = "/usr/local/bin/rollapp_evm"
	}
	return rollappBinaryPath
}
