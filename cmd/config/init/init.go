package initconfig

import (
	"fmt"
	"os"

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
		Use:   "init <chain-id> <denom>",
		Short: "Initialize a RollApp configuration on your local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig := getInitConfig(cmd, args)
			isUniqueRollapp, err := isRollappIDUnique(initConfig.RollappID)
			OutputCleanError(err)
			if !isUniqueRollapp {
				OutputCleanError(fmt.Errorf("Rollapp ID %s already exists on the hub. Please use a unique ID.", initConfig.RollappID))
			}
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
				AddressPrefix: addressPrefixes.Rollapp,
			}, ChainConfig{
				ID:            HubData.ID,
				RPC:           cmd.Flag(FlagNames.HubRPC).Value.String(),
				Denom:         "udym",
				AddressPrefix: addressPrefixes.Hub,
			}, initConfig))
			printInitOutput(addresses, initConfig.RollappID)
		},
		Args: cobra.ExactArgs(2),
	}

	addFlags(cmd)
	return cmd
}
