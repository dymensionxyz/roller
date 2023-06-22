package initconfig

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init <rollapp-id> <denom>",
		Short: "Initialize a RollApp configuration on your local machine.",
		Long: fmt.Sprintf(`Initialize a RollApp configuration on your local machine.
		
%s
`, getValidRollappIdMessage()),
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
					utils.PrettifyErrorIfExists(os.MkdirAll(initConfig.Home, 0755))
				} else {
					os.Exit(0)
				}
			} else {
				utils.PrettifyErrorIfExists(os.MkdirAll(initConfig.Home, 0755))
			}
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
			addresses, err := generateKeys(initConfig)
			utils.PrettifyErrorIfExists(err)
			utils.PrettifyErrorIfExists(initializeLightNodeConfig(initConfig))
			daAddress, err := utils.GetCelestiaAddress(filepath.Join(initConfig.Home, consts.ConfigDirName.DALightNode, consts.KeysDirName))
			utils.PrettifyErrorIfExists(err)
			addresses[consts.KeyNames.DALightNode] = daAddress
			initializeRollappConfig(initConfig)
			utils.PrettifyErrorIfExists(initializeRollappGenesis(initConfig))
			utils.PrettifyErrorIfExists(utils.WriteConfigToTOML(initConfig))
			printInitOutput(addresses, initConfig.RollappID)
		},
		Args: cobra.ExactArgs(2),
	}
	utils.AddGlobalFlags(initCmd)
	addFlags(initCmd)
	return initCmd
}
