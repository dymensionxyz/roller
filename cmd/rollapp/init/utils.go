package initrollapp

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	global_utils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/archives"
)

func expandHomePath(path string) (string, error) {
	if path[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return "", err
		}
		path = filepath.Join(usr.HomeDir, path[2:])
	}
	return path, nil
}

// in runInit I parse the entire genesis creator zip file twice to extract
// the file this looks awful but since the archive has only 2 files it's
// kinda fine
func runInit(cmd *cobra.Command, configArchivePath string) error {
	home := utils.GetRollerRootDir()
	outputHandler := initconfig.NewOutputHandler(false)

	defer outputHandler.StopSpinner()

	isRootExist, err := global_utils.DirNotEmpty(home)
	if err != nil {
		utils.PrettifyErrorIfExists(err)
		return err
	}

	if isRootExist {
		outputHandler.StopSpinner()
		shouldOverwrite, err := outputHandler.PromptOverwriteConfig(home)
		if err != nil {
			utils.PrettifyErrorIfExists(err)
			fmt.Println("prompt")
			return err
		}
		if shouldOverwrite {
			err = os.RemoveAll(home)
			if err != nil {
				fmt.Println("RemoveAll")
				utils.PrettifyErrorIfExists(err)
				return err
			}
		} else {
			os.Exit(0)
		}
	}

	// nolint:gofumpt
	err = os.MkdirAll(home, 0755)
	if err != nil {
		utils.PrettifyErrorIfExists(err)
		return err
	}

	err = archives.ExtractFileFromNestedTar(
		configArchivePath,
		config.RollerConfigFileName,
		home,
	)
	if err != nil {
		fmt.Println("nested tar")
		return err
	}

	initConfigPtr, err := initconfig.GetInitConfig(cmd)
	if err != nil {
		fmt.Println("getiniconfig")
		utils.PrettifyErrorIfExists(err)
		return err
	}

	initConfig := *initConfigPtr

	utils.RunOnInterrupt(outputHandler.StopSpinner)
	outputHandler.StartSpinner(consts.SpinnerMsgs.UniqueIdVerification)
	err = initConfig.Validate()
	if err != nil {
		fmt.Println("validate")
		utils.PrettifyErrorIfExists(err)
		return err
	}

	// TODO: create all dirs here
	outputHandler.StartSpinner(" Initializing RollApp configuration files...")
	/* ---------------------------- Initialize relayer --------------------------- */
	// 20240607 relayer will be handled using a separate, relayer command
	// rollerConfigFilePath := filepath.Join(utils.GetRollerRootDir(), "roller.toml")
	// rollappPrefix, err := global_utils.GetKeyFromTomlFile(rollerConfigFilePath, "bech32_prefix")
	// utils.PrettifyErrorIfExists(err)

	// err = initializeRelayerConfig(relayer.ChainConfig{
	// 	ID:            initConfig.RollappID,
	// 	RPC:           consts.DefaultRollappRPC,
	// 	Denom:         initConfig.Denom,
	// 	AddressPrefix: rollappPrefix,
	// 	GasPrices:     "0",
	// }, relayer.ChainConfig{
	// 	ID:            initConfig.HubData.ID,
	// 	RPC:           initConfig.HubData.ARCHIVE_RPC_URL,
	// 	Denom:         consts.Denoms.Hub,
	// 	AddressPrefix: consts.AddressPrefixes.Hub,
	// 	GasPrices:     initConfig.HubData.GAS_PRICE,
	// }, initConfig)
	// if err != nil {
	// 	return err
	// }

	/* ------------------------------ Generate keys ----------------------------- */
	addresses, err := initconfig.GenerateKeys(initConfig)
	if err != nil {
		utils.PrettifyErrorIfExists(err)
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
	err = initconfig.InitializeRollappConfig(initConfig)
	if err != nil {
		return err
	}

	// genesis creator archive
	err = archives.ExtractFileFromNestedTar(
		configArchivePath,
		"genesis.json",
		filepath.Join(home, consts.ConfigDirName.Rollapp, "config"),
	)
	if err != nil {
		return err
	}

	// adds the sequencer address to the whitelists
	err = initconfig.UpdateGenesisParams(home)
	if err != nil {
		return err
	}

	rollerConfigFilePath := filepath.Join(utils.GetRollerRootDir(), "roller.toml")
	err = global_utils.UpdateFieldInToml(rollerConfigFilePath, "home", utils.GetRollerRootDir())
	if err != nil {
		fmt.Println("failed to add home to roller.toml: ", err)
		return err
	}

	// 20240607 genesis is generated using the genesis-creator
	// err = initializeRollappGenesis(initConfig)
	// if err != nil {
	// 	return err
	// }

	// TODO: review, roller config is generated using genesis-creator
	// some of the config values should be moved there
	// err = config.WriteConfigToTOML(initConfig)
	// if err != nil {
	// 	return err
	// }

	/* ------------------------------ Initialize Local Hub ---------------------------- */
	// TODO: local hub is out of scope, implement as the last step
	// hub := cmd.Flag(FlagNames.HubID).Value.String()
	// if hub == consts.LocalHubName {
	// 	err := initLocalHub(initConfig)
	// 	utils.PrettifyErrorIfExists(err)
	// }

	/* ------------------------------ Print output ------------------------------ */
	outputHandler.StopSpinner()
	outputHandler.PrintInitOutput(initConfig, addresses, initConfig.RollappID)

	return nil
}

func checkConfigArchive(path string) (string, error) {
	path = strings.TrimSpace(path)

	if path == "" {
		return "", errors.New("invalid or no input")
	}

	archivePath, err := expandHomePath(path)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return "", err
	}

	return archivePath, nil
}
