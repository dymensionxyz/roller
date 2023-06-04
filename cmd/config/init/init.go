package initconfig

import (
	"github.com/spf13/cobra"
)

type InitConfig struct {
	Home              string
	RollappID         string
	RollappBinary     string
	CreateDALightNode bool
	Denom             string
	HubID             string
	Decimals          uint64
}

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <chain-id>",
		Short: "Initialize a rollapp configuration on your local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			rollappId := args[0]
			denom := args[1]
			home := cmd.Flag(FlagNames.Home).Value.String()
			createLightNode := !cmd.Flags().Changed(lightNodeEndpointFlag)
			rollappBinaryPath := getRollappBinaryPath(cmd)
			decimals := getDecimals(cmd)
			initConfig := InitConfig{
				Home:              home,
				RollappID:         rollappId,
				RollappBinary:     rollappBinaryPath,
				CreateDALightNode: createLightNode,
				Denom:             denom,
				HubID:             DefaultHubID,
				Decimals:          decimals,
			}

			addresses := initializeKeys(initConfig)
			if createLightNode {
				if err := initializeLightNodeConfig(initConfig); err != nil {
					panic(err)
				}
			}
			initializeRollappConfig(initConfig)
			if err := initializeRollappGenesis(initConfig); err != nil {
				panic(err)
			}

			if err := initializeRelayerConfig(ChainConfig{
				ID:            rollappId,
				RPC:           defaultRollappRPC,
				Denom:         denom,
				AddressPrefix: addressPrefixes.Rollapp,
			}, ChainConfig{
				ID:            DefaultHubID,
				RPC:           cmd.Flag(FlagNames.HubRPC).Value.String(),
				Denom:         "udym",
				AddressPrefix: addressPrefixes.Hub,
			}, initConfig); err != nil {
				panic(err)
			}
			celestiaAddress := addresses[KeyNames.DALightNode]
			rollappHubAddress := addresses[KeyNames.HubSequencer]
			relayerHubAddress := addresses[KeyNames.HubRelayer]
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
	cmd.Flags().StringP(FlagNames.HubRPC, "", defaultHubRPC, "Dymension Hub rpc endpoint")
	cmd.Flags().StringP(FlagNames.LightNodeEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided")
	cmd.Flags().StringP(FlagNames.RollappBinary, "", "", "The rollapp binary. Should be passed only if you built a custom rollapp")
	cmd.Flags().Uint64P(FlagNames.Decimals, "", 18, "The number of decimal places a rollapp token supports")
	cmd.Flags().StringP(FlagNames.Home, "", getRollerRootDir(), "The directory of the roller config files")
}

func getDecimals(cmd *cobra.Command) uint64 {
	decimals, err := cmd.Flags().GetUint64(FlagNames.Decimals)
	if err != nil {
		panic(err)
	}
	return decimals
}

func initializeKeys(initConfig InitConfig) map[string]string {
	if initConfig.CreateDALightNode {
		addresses, err := generateKeys(initConfig)
		if err != nil {
			panic(err)
		}
		return addresses
	} else {
		addresses, err := generateKeys(initConfig, KeyNames.DALightNode)
		if err != nil {
			panic(err)
		}
		return addresses
	}
}

func getRollappBinaryPath(cmd *cobra.Command) string {
	rollappBinaryPath := cmd.Flag(FlagNames.RollappBinary).Value.String()
	if rollappBinaryPath == "" {
		return defaultRollappBinaryPath
	}
	return rollappBinaryPath
}
