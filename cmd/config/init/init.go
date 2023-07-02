package initconfig

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:   "init <chain-id> <denom>",
		Short: "Initialize a RollApp configuration on your local machine.",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			interactive, _ := cmd.Flags().GetBool(FlagNames.Interactive)
			if interactive {
				return nil
			}

			if len(args) < 2 {
				return fmt.Errorf("invalid number of arguments. Expected 2, got %d", len(args))
			}

			//TODO: parse the config here instead of GetInitConfig in Run command
			// cmd.SetContextValue("mydata", data)

			return nil
		},
		Run: func(cmd *cobra.Command, args []string) {
			initConfig, err := GetInitConfig(cmd, args)
			utils.PrettifyErrorIfExists(err)

			err = initConfig.Validate()
			err = errors.Wrap(err, getValidRollappIdMessage())
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
	}

	addFlags(initCmd)
	return initCmd
}
