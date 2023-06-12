package initconfig

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/dymensionxyz/roller/cmd/utils"
)

type InitConfig struct {
	Home              string
	RollappID         string
	RollappBinary     string
	CreateDALightNode bool
	Denom             string
	Decimals          uint64
	HubID             string
}

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <chain-id> <denom>",
		Short: "Initialize a RollApp configuration on your local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig := GetInitConfig(cmd, args)
			utils.PrettifyErrorIfExists(VerifyUniqueRollappID(initConfig.RollappID))
			isRootExist, err := dirNotEmpty(initConfig.Home)
			utils.PrettifyErrorIfExists(err)
			if isRootExist {
				shouldOverwrite, err := promptOverwriteConfig(initConfig.Home)
				utils.PrettifyErrorIfExists(err)
				if shouldOverwrite {
					utils.PrettifyErrorIfExists(os.RemoveAll(initConfig.Home))
				} else {
					os.Exit(0)
				}
			}

			addresses := initializeKeys(initConfig)
			if initConfig.CreateDALightNode {
				utils.PrettifyErrorIfExists(initializeLightNodeConfig(initConfig))
			}
			initializeRollappConfig(initConfig)
			utils.PrettifyErrorIfExists(initializeRollappGenesis(initConfig))
			utils.PrettifyErrorIfExists(initializeRelayerConfig(ChainConfig{
				ID:            initConfig.RollappID,
				RPC:           defaultRollappRPC,
				Denom:         initConfig.Denom,
				AddressPrefix: AddressPrefixes.Rollapp,
			}, ChainConfig{
				ID:            HubData.ID,
				RPC:           cmd.Flag(FlagNames.HubRPC).Value.String(),
				Denom:         "udym",
				AddressPrefix: AddressPrefixes.Hub,
			}, initConfig))
			utils.PrettifyErrorIfExists(WriteConfigToTOML(initConfig))
			printInitOutput(addresses, initConfig.RollappID)
		},
		Args: cobra.ExactArgs(2),
	}

	addFlags(cmd)
	return cmd
}
