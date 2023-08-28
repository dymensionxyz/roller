package initconfig

import (
	"fmt"
	global_utils "github.com/dymensionxyz/roller/utils"
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

A valid denom should consist of 3-6 English alphabet letters, for example, 'btc', 'eth', 'pepe', etc.`
}

func runInit(cmd *cobra.Command, args []string) error {
	noOutput, err := cmd.Flags().GetBool(FlagNames.NoOutput)
	if err != nil {
		return err
	}
	initConfig, err := GetInitConfig(cmd, args)
	if err != nil {
		return err
	}
	outputHandler := NewOutputHandler(noOutput)
	defer outputHandler.StopSpinner()
	utils.RunOnInterrupt(outputHandler.StopSpinner)
	outputHandler.StartSpinner(consts.SpinnerMsgs.UniqueIdVerification)
	err = initConfig.Validate()
	if err != nil {
		return err
	}
	err = VerifyUniqueRollappID(initConfig.RollappID, initConfig)
	if err != nil {
		return err
	}
	isRootExist, err := global_utils.DirNotEmpty(initConfig.Home)
	if err != nil {
		return err
	}
	if isRootExist {
		outputHandler.StopSpinner()
		shouldOverwrite, err := outputHandler.PromptOverwriteConfig(initConfig.Home)
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
	}
	err = os.MkdirAll(initConfig.Home, 0755)
	if err != nil {
		return err
	}
	//TODO: create all dirs here
	outputHandler.StartSpinner(" Initializing RollApp configuration files...")
	/* ---------------------------- Initialize relayer --------------------------- */
	rollappPrefix, err := utils.GetAddressPrefix(initConfig.RollappBinary)
	utils.PrettifyErrorIfExists(err)
	err = initializeRelayerConfig(ChainConfig{
		ID:            initConfig.RollappID,
		RPC:           consts.DefaultRollappRPC,
		Denom:         initConfig.Denom,
		AddressPrefix: rollappPrefix,
		GasPrices:     "0",
	}, ChainConfig{
		ID:            initConfig.HubData.ID,
		RPC:           initConfig.HubData.RPC_URL,
		Denom:         consts.Denoms.Hub,
		AddressPrefix: consts.AddressPrefixes.Hub,
		GasPrices:     initConfig.HubData.GAS_PRICE,
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
			Name: damanager.GetKeyName(),
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

	/* ------------------------------ Initialize Local Hub ---------------------------- */
	hub := cmd.Flag(FlagNames.HubID).Value.String()
	if hub == LocalHubName {
		err := initLocalHub(initConfig)
		utils.PrettifyErrorIfExists(err)
	}

	/* ------------------------------ Print output ------------------------------ */
	outputHandler.StopSpinner()
	outputHandler.printInitOutput(initConfig, addresses, initConfig.RollappID)

	return nil
}
