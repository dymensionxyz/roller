package initconfig

import (
	"os"

	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/spf13/cobra"
)

type InitConfig struct {
	Home          string
	RollappID     string
	RollappBinary string
	Denom         string
	Decimals      uint64
	HubData       HubData
}

func InitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init <chain-id> <denom>",
		Short: "Initialize a RollApp configuration on your local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			initConfig := GetInitConfig(cmd, args)
			utils.PrettifyErrorIfExists(VerifyUniqueRollappID(initConfig.RollappID, initConfig))
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
			addresses, err := generateKeys(initConfig)
			utils.PrettifyErrorIfExists(err)
			utils.PrettifyErrorIfExists(initializeLightNodeConfig(initConfig))
			initializeRollappConfig(initConfig)
			utils.PrettifyErrorIfExists(initializeRollappGenesis(initConfig))
			utils.PrettifyErrorIfExists(initializeRelayerConfig(ChainConfig{
				ID:            initConfig.RollappID,
				RPC:           defaultRollappRPC,
				Denom:         initConfig.Denom,
				AddressPrefix: consts.AddressPrefixes.Rollapp,
			}, ChainConfig{
				ID:            initConfig.HubData.ID,
				RPC:           initConfig.HubData.RPC_URL,
				Denom:         "udym",
				AddressPrefix: consts.AddressPrefixes.Hub,
			}, initConfig))
			utils.PrettifyErrorIfExists(WriteConfigToTOML(initConfig))
			daLightNodeAddress, err := utils.GetCelestiaAddress(filepath.Join(initConfig.Home, consts.ConfigDirName.DALightNode, KeysDirName))
			utils.PrettifyErrorIfExists(err)
			addresses[consts.KeyNames.DALightNode] = daLightNodeAddress
			printInitOutput(addresses, initConfig.RollappID)
		},
		Args: cobra.ExactArgs(2),
	}
	utils.AddGlobalFlags(initCmd)
	addFlags(initCmd)
	return initCmd
}
