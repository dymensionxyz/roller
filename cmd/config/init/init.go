package initconfig

import (
	"fmt"
	"os"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

func InitCmd() *cobra.Command {
	initCmd := &cobra.Command{
		Use:     "init <rollapp-id> <denom> | --interactive",
		Short:   "Initialize a RollApp configuration on your local machine.",
		Long:    "Initialize a RollApp configuration on your local machine\n" + requiredFlagsUsage(),
		Example: `init mars_9721-1 btc`,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			interactive, _ := cmd.Flags().GetBool(FlagNames.Interactive)
			if interactive {
				return nil
			}

			if len(args) == 0 {
				fmt.Println("No arguments provided. Running in interactive mode.")
				if err := cmd.Flags().Set(FlagNames.Interactive, "true"); err != nil {
					return err
				}
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
			spin := utils.GetLoadingSpinner()
			spin.Suffix = consts.SpinnerMsgs.UniqueIdVerification
			spin.Start()
			initConfig, err := GetInitConfig(cmd, args)
			utils.PrettifyErrorIfExists(err)

			err = initConfig.Validate()
			utils.PrettifyErrorIfExists(err, func() {
				fmt.Println(requiredFlagsUsage())
			})

			utils.PrettifyErrorIfExists(VerifyUniqueRollappID(initConfig.RollappID, initConfig))
			isRootExist, err := dirNotEmpty(initConfig.Home)
			utils.PrettifyErrorIfExists(err)
			if isRootExist {
				spin.Stop()
				shouldOverwrite, err := promptOverwriteConfig(initConfig.Home)
				utils.PrettifyErrorIfExists(err)
				if shouldOverwrite {
					utils.PrettifyErrorIfExists(os.RemoveAll(initConfig.Home))
				} else {
					os.Exit(0)
				}
				spin.Start()
			}
			utils.PrettifyErrorIfExists(os.MkdirAll(initConfig.Home, 0755))

			//TODO: create all dirs here
			spin.Suffix = " Initializing RollApp configuration files..."
			spin.Restart()
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
			utils.PrettifyErrorIfExists(config.WriteConfigToTOML(initConfig))

			/* ------------------------------ Print output ------------------------------ */
			spin.Stop()
			printInitOutput(initConfig, addresses, initConfig.RollappID)
		},
	}

	if err := addFlags(initCmd); err != nil {
		panic(err)
	}

	return initCmd
}

func requiredFlagsUsage() string {
	return `
A valid RollApp ID should follow the format 'name_uniqueID-revision', where
- 'name' is made up of lowercase English letters
- 'uniqueID' is a 1 to 5 digit number representing the unique ID EIP155 rollapp ID
- 'revision' is a 1 to 5 digit number representing the revision number for this rollapp

A valid denom should consist of exactly 3 English alphabet letters, for example 'btc', 'eth'`
}
