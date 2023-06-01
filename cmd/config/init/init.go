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
			rollappKeyPrefix := getKeyPrefix(cmd.Flag(flagNames.KeyPrefix).Value.String(), rollappId)
			createLightNode := !cmd.Flags().Changed(lightNodeEndpointFlag)
			addresses := initializeKeys(rollappId, defaultHubId, rollappKeyPrefix, createLightNode)
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

			if err := initializeRelayerConfig(ChainConfig{
				ID:        rollappId,
				RPC:       defaultRollappRPC,
				Denom:     denom,
				KeyPrefix: rollappKeyPrefix,
			}, ChainConfig{
				ID:        defaultHubId,
				RPC:       cmd.Flag(flagNames.HubRPC).Value.String(),
				Denom:     "udym",
				KeyPrefix: keyPrefixes.Hub,
			}); err != nil {
				panic(err)
			}
			celestiaAddress := addresses[keyNames.DALightNode]
			rollappHubAddress := addresses[keyNames.HubSequencer]
			relayerHubAddress := addresses[keyNames.HubRelayer]
			printInitOutput(AddressesToFund{
				DA:           celestiaAddress,
				HubSequencer: rollappHubAddress,
				HubRelayer:   relayerHubAddress,
			}, rollappId)
		},
		Args: cobra.ExactArgs(2),
	}

	addFlags(cmd)
	return cmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().StringP(flagNames.HubRPC, "", defaultHubRPC, "Dymension Hub rpc endpoint")
	cmd.Flags().StringP(flagNames.LightNodeEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided")
	cmd.Flags().StringP(flagNames.KeyPrefix, "", "", "The `bech32` prefix of the rollapp keys. Defaults to the first three characters of the chain-id")
	cmd.Flags().StringP(flagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().Uint64P(flagNames.Decimals, "", 18, "The number of decimal places a rollapp token supports")
	cmd.Flags().StringP(flagNames.Home, "", getRollerRootDir(), "The directory of the roller config files")
}

func initializeKeys(rollappId, hubId, rollappPrefix string, createLightNode bool) map[string]string {
	if createLightNode {
		addresses, err := generateKeys(rollappId, defaultHubId, rollappPrefix)
		if err != nil {
			panic(err)
		}
		return addresses
	} else {
		addresses, err := generateKeys(rollappId, defaultHubId, rollappPrefix, keyNames.DALightNode)
		if err != nil {
			panic(err)
		}
		return addresses
	}
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
