package init

import (
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <chain-id>",
		Short: "Initialize a rollapp configuration on your local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			rollappId := args[0]
			denom := args[1]
			createLightNode := !cmd.Flags().Changed(lightNodeEndpointFlag)
			if createLightNode {
				if err := generateKeys(rollappId, defaultHubId); err != nil {
					panic(err)
				}
			} else {
				if err := generateKeys(rollappId, defaultHubId, keyNames.DALightNode); err != nil {
					panic(err)
				}
			}
			rollappBinaryPath := getRollappBinaryPath(cmd.Flag(flagNames.RollappBinary).Value.String())
			decimals, err := cmd.Flags().GetUint64(flagNames.Decimals)
			if err != nil {
				panic(err)
			}
			if createLightNode {
				if err = initializeLightNodeConfig(); err != nil {
					panic(err)
				}
			}
			initializeRollappConfig(rollappBinaryPath, rollappId, denom)
			if err = initializeRollappGenesis(rollappBinaryPath, decimals, denom); err != nil {
				panic(err)
			}

			rollappKeyPrefix := getKeyPrefix(cmd.Flag(flagNames.KeyPrefix).Value.String(), rollappId)
			if err := initializeRelayerConfig(ChainConfig{
				ID:        rollappId,
				RPC:       defaultRollappRPC,
				Denom:     denom,
				KeyPrefix: rollappKeyPrefix,
			}, ChainConfig{
				ID:        defaultHubId,
				RPC:       cmd.Flag(flagNames.HubRPC).Value.String(),
				Denom:     "udym",
				KeyPrefix: "dym",
			}); err != nil {
				panic(err)
			}
		},
		Args: cobra.ExactArgs(2),
	}
	cmd.Flags().StringP(flagNames.HubRPC, "", defaultHubRPC, "Dymension Hub rpc endpoint")
	cmd.Flags().StringP(flagNames.LightNodeEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided.")
	cmd.Flags().StringP(flagNames.KeyPrefix, "", "", "The `bech32` prefix of the rollapp keys. Defaults to the first three characters of the chain-id.")
	cmd.Flags().StringP(flagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp.")
	cmd.Flags().Uint64P(flagNames.Decimals, "", 18, "The number of decimal places a rollapp token supports.")
	return cmd
}

func getRollappBinaryPath(rollappBinaryPath string) string {
	if rollappBinaryPath == "" {
		return defaultRollappBinaryPath
	}
	return rollappBinaryPath
}

func getKeyPrefix(keyPrefix string, chainId string) string {
	if keyPrefix == "" {
		return chainId[:3]
	}
	return keyPrefix
}
