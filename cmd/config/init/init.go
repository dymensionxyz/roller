package initconfig

import (
	"fmt"
	"os"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init <chain-id> <denom>",
		Short: "Initialize a RollApp configuration on your local machine.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			err := verifyHubID(cmd)
			if err != nil {
				return err
			}
			err = verifyTokenSupply(cmd)
			if err != nil {
				return err
			}
			rollappID := args[0]
			if !validateRollAppID(rollappID) {
				return fmt.Errorf("invalid RollApp ID '%s'. %s", rollappID, getValidRollappIdMessage())
			}
			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			initConfig, err := GetInitConfig(cmd, args)
			utils.PrettifyErrorIfExists(err)
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
			utils.PrettifyErrorIfExists(os.MkdirAll(initConfig.Home, 0755))

			//TODO: create all dirs here

			/* ---------------------------- Initilize relayer --------------------------- */
			utils.PrettifyErrorIfExists(initializeRelayerConfig(ChainConfig{
				ID:            initConfig.RollappID,
				RPC:           consts.DefaultRollappRPC,
				Denom:         initConfig.Denom,
				AddressPrefix: consts.AddressPrefixes.Rollapp,
			}, ChainConfig{
				ID:            initConfig.HubData.ID,
				RPC:           initConfig.HubData.RPC_URL,
				Denom:         consts.Denoms.Hub,
				AddressPrefix: consts.AddressPrefixes.Hub,
			}, initConfig))

			/* ------------------------------ Generate keys ----------------------------- */
			addresses, err := generateKeys(initConfig)
			utils.PrettifyErrorIfExists(err)

			/* ------------------------ Initialize DA light node ------------------------ */
			utils.PrettifyErrorIfExists(initializeLightNodeConfig(initConfig))
			daAddress, err := utils.GetCelestiaAddress(initConfig.Home)
			utils.PrettifyErrorIfExists(err)
			addresses = append(addresses, utils.AddressData{
				Addr: daAddress,
				Name: consts.KeysIds.DALightNode,
			})

			/* --------------------------- Initiailize Rollapp -------------------------- */
			utils.PrettifyErrorIfExists(initializeRollappConfig(initConfig))
			utils.PrettifyErrorIfExists(initializeRollappGenesis(initConfig))
			utils.PrettifyErrorIfExists(utils.WriteConfigToTOML(initConfig))

			/* ------------------------------ Print output ------------------------------ */
			printInitOutput(addresses, initConfig.RollappID)
		},
		Args: cobra.ExactArgs(2),
	}
	utils.AddGlobalFlags(initCmd)
	addFlags(initCmd)
	return initCmd
}
