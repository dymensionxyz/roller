package initconfig

import (
	"os"

	"github.com/spf13/cobra"
)

type InitConfig struct {
	Home              string
	RollappID         string
	RollappBinary     string
	CreateDALightNode bool
	Denom             string
	Decimals          uint64
	HubData           HubData
}

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <chain-id> <denom>",
		Short: "Initialize a RollApp configuration on your local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig := GetInitConfig(cmd, args)
			OutputCleanError(VerifyUniqueRollappID(initConfig.RollappID, initConfig))
			isRootExist, err := dirNotEmpty(initConfig.Home)
			OutputCleanError(err)
			if isRootExist {
				shouldOverwrite, err := promptOverwriteConfig(initConfig.Home)
				OutputCleanError(err)
				if shouldOverwrite {
					OutputCleanError(os.RemoveAll(initConfig.Home))
				} else {
					os.Exit(0)
				}
			}

			addresses := initializeKeys(initConfig)
			if initConfig.CreateDALightNode {
				OutputCleanError(initializeLightNodeConfig(initConfig))
			}
			initializeRollappConfig(initConfig)
			OutputCleanError(initializeRollappGenesis(initConfig))
			OutputCleanError(initializeRelayerConfig(ChainConfig{
				ID:            initConfig.RollappID,
				RPC:           defaultRollappRPC,
				Denom:         initConfig.Denom,
				AddressPrefix: AddressPrefixes.Rollapp,
			}, ChainConfig{
				ID:            initConfig.HubData.ID,
				RPC:           initConfig.HubData.RPC_URL,
				Denom:         "udym",
				AddressPrefix: AddressPrefixes.Hub,
			}, initConfig))
			OutputCleanError(WriteConfigToTOML(initConfig))
			printInitOutput(addresses, initConfig.RollappID)
		},
		Args: cobra.ExactArgs(2),
	}

	addFlags(cmd)
	return cmd
}
