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
			initConfig := getInitConfig(cmd, args)
			overwrite, err := prepareDirectory(initConfig.Home)
			handleError(err)
			if overwrite {
				addresses := initializeKeys(initConfig)
				if initConfig.CreateDALightNode {
					handleError(initializeLightNodeConfig(initConfig))
				}
				initializeRollappConfig(initConfig)
				handleError(initializeRollappGenesis(initConfig))
				handleError(initializeRelayerConfig(ChainConfig{
					ID:            initConfig.RollappID,
					RPC:           defaultRollappRPC,
					Denom:         initConfig.Denom,
					AddressPrefix: addressPrefixes.Rollapp,
				}, ChainConfig{
					ID:            DefaultHubID,
					RPC:           cmd.Flag(FlagNames.HubRPC).Value.String(),
					Denom:         "udym",
					AddressPrefix: addressPrefixes.Hub,
				}, initConfig))
				printInitOutput(addresses, initConfig.RollappID)
			}
		},
		Args: cobra.ExactArgs(2),
	}

	addFlags(cmd)
	return cmd
}
