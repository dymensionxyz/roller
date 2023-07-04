package initconfig

import (
	"fmt"
	"os"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
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
			utils.PrettifyErrorIfExists(runInit(cmd, args))
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
- 'uniqueID' is a number up to the length of 5 digits representing the unique ID EIP155 rollapp ID
- 'revision' is a number up to the length of 5 digits representing the revision number for this rollapp

A valid denom should consist of exactly 3 English alphabet letters, for example 'btc', 'eth'`
}

func runInit(cmd *cobra.Command, args []string) error {
	initConfig, err := GetInitConfig(cmd, args)
	if err != nil {
		return err
	}
	spin := utils.GetLoadingSpinner()
	spin.Suffix = consts.SpinnerMsgs.UniqueIdVerification
	spin.Start()
	defer spin.Stop()
	err = initConfig.Validate()
	if err != nil {
		return err
	}
	err = VerifyUniqueRollappID(initConfig.RollappID, initConfig)
	if err != nil {
		return err
	}
	isRootExist, err := dirNotEmpty(initConfig.Home)
	if err != nil {
		return err
	}
	if isRootExist {
		spin.Stop()
		shouldOverwrite, err := promptOverwriteConfig(initConfig.Home)
		if err != nil {
			return err
		}
		if shouldOverwrite {
			err = os.RemoveAll(initConfig.Home)
			if err != nil {
				return err
			}
		} else {
			os.Exit(0)
		}
		spin.Start()
	}
	err = os.MkdirAll(initConfig.Home, 0755)
	if err != nil {
		return err
	}
	//TODO: create all dirs here
	spin.Suffix = " Initializing RollApp configuration files..."
	spin.Restart()
	/* ---------------------------- Initialize relayer --------------------------- */
	err = initializeRelayerConfig(ChainConfig{
		ID:            initConfig.RollappID,
		RPC:           consts.DefaultRollappRPC,
		Denom:         initConfig.Denom,
		AddressPrefix: consts.AddressPrefixes.Rollapp,
	}, ChainConfig{
		ID:            initConfig.HubData.ID,
		RPC:           initConfig.HubData.RPC_URL,
		Denom:         consts.Denoms.Hub,
		AddressPrefix: consts.AddressPrefixes.Hub,
	}, initConfig)
	if err != nil {
		return err
	}

	/* ------------------------------ Generate keys ----------------------------- */
	addresses, err := generateKeys(initConfig)
	if err != nil {
		return err
	}

	/* ------------------------ Initialize DA light node ------------------------ */
	damanager := datalayer.NewDAManager(initConfig.DA, initConfig.Home)
	err = damanager.InitializeLightNodeConfig()
	if err != nil {
		return err
	}

	daAddress, err := damanager.GetDAAccountAddress()
	if err != nil {
		return err
	}

	if daAddress != "" {
		addresses = append(addresses, utils.AddressData{
			Name: consts.KeysIds.DALightNode,
			Addr: daAddress,
		})
	}

	/* --------------------------- Initialize Rollapp -------------------------- */
	err = initializeRollappConfig(initConfig)
	if err != nil {
		return err
	}

	err = initializeRollappGenesis(initConfig)
	if err != nil {
		return err
	}

	err = config.WriteConfigToTOML(initConfig)
	if err != nil {
		return err
	}

	/* ------------------------------ Print output ------------------------------ */
	spin.Stop()
	printInitOutput(initConfig, addresses, initConfig.RollappID)

	return nil
}
