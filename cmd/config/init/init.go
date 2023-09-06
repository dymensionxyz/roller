package initconfig

import (
	"fmt"
	"github.com/dymensionxyz/roller/relayer"
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
		Example: `init mars btc`,
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

const validRollappIDMsg = "A valid RollApp ID should contain only lowercase alphabetical characters, for example, 'mars', 'venus', 'earth', etc."

func requiredFlagsUsage() string {
	return fmt.Sprintf(`
%s

A valid denom should consist of 3-6 English alphabet letters, for example, 'btc', 'eth', 'pepe', etc.`, validRollappIDMsg)
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
	err = initializeRelayerConfig(relayer.ChainConfig{
		ID:            initConfig.RollappID,
		RPC:           consts.DefaultRollappRPC,
		Denom:         initConfig.Denom,
		AddressPrefix: rollappPrefix,
		GasPrices:     "0",
	}, relayer.ChainConfig{
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
