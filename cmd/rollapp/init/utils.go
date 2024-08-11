package initrollapp

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	cmdutils "github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config/toml"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

// in runInit I parse the entire genesis creator zip file twice to extract
// the file this looks awful but since the archive has only 2 files it's
// kinda fine
func runInit(cmd *cobra.Command, env string, raID string) error {
	home, err := globalutils.ExpandHomePath(cmd.Flag(cmdutils.FlagNames.Home).Value.String())
	if err != nil {
		pterm.Error.Println("failed to expand home directory")
		return err
	}

	outputHandler := initconfig.NewOutputHandler(false)

	defer outputHandler.StopSpinner()

	// TODO: extract into util
	isRootExist, err := globalutils.DirNotEmpty(home)
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return err
	}

	if isRootExist {
		outputHandler.StopSpinner()
		shouldOverwrite, err := outputHandler.PromptOverwriteConfig(home)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return err
		}
		if shouldOverwrite {
			err = os.RemoveAll(home)
			if err != nil {
				errorhandling.PrettifyErrorIfExists(err)
				return err
			}
			// nolint:gofumpt
			err = os.MkdirAll(home, 0o755)
			if err != nil {
				errorhandling.PrettifyErrorIfExists(err)
				return err
			}
		} else {
			os.Exit(0)
		}
	}

	// initConfigPtr, err := initconfig.GetInitConfig(cmd, options.useMockSettlement)

	initConfigPtr, err := toml.LoadRollappMetadataFromChain(home, raID)
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return err
	}
	initConfig := *initConfigPtr

	errorhandling.RunOnInterrupt(outputHandler.StopSpinner)
	outputHandler.StartSpinner(consts.SpinnerMsgs.UniqueIdVerification)
	err = initConfig.Validate()
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return err
	}

	// TODO: create all dirs here
	outputHandler.StartSpinner(" Initializing RollApp configuration files...\n")

	/* ------------------------------ Generate keys ----------------------------- */
	addresses, err := initconfig.GenerateSequencersKeys(initConfig)
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return err
	}

	/* ------------------------ Initialize DA light node ------------------------ */
	damanager := datalayer.NewDAManager(initConfig.DA, initConfig.Home)
	mnemonic, err := damanager.InitializeLightNodeConfig()
	if err != nil {
		return err
	}

	daAddress, err := damanager.GetDAAccountAddress()
	if err != nil {
		return err
	}

	if daAddress != nil {
		addresses = append(
			addresses, utils.KeyInfo{
				Name:     damanager.GetKeyName(),
				Address:  daAddress.Address,
				Mnemonic: mnemonic,
			},
		)
	}

	/* --------------------------- Initialize Rollapp -------------------------- */
	err = initconfig.InitializeRollappConfig(initConfig)
	if err != nil {
		return err
	}

	// TODO: where to retrieve genesis from?
	// if configArchivePath != "" {
	// 	// genesis creator archive
	// 	err = archives.ExtractFileFromNestedTar(
	// 		configArchivePath,
	// 		"genesis.json",
	// 		filepath.Join(home, consts.ConfigDirName.Rollapp, "config"),
	// 	)
	// 	if err != nil {
	// 		return err
	// 	}
	// }

	// adds the sequencer address to the whitelists
	err = initconfig.UpdateGenesisParams(home, &initConfig)
	if err != nil {
		return err
	}

	rollerConfigFilePath := filepath.Join(home, "roller.toml")
	err = globalutils.UpdateFieldInToml(rollerConfigFilePath, "home", home)
	if err != nil {
		fmt.Println("failed to add home to roller.toml: ", err)
		return err
	}

	hd := consts.Hubs[env]
	rollerTomlData := map[string]string{
		"HubData.ID":              hd.ID,
		"HubData.api_url":         hd.API_URL,
		"HubData.rpc_url":         hd.RPC_URL,
		"HubData.archive_rpc_url": hd.ARCHIVE_RPC_URL,
		"HubData.gas_price":       hd.GAS_PRICE,
		"da":                      strings.ToLower(string(initConfig.DA)),
		"rollapp_binary":          strings.ToLower(consts.Executables.RollappEVM),
	}

	for key, value := range rollerTomlData {
		err = globalutils.UpdateFieldInToml(
			rollerConfigFilePath,
			key,
			value,
		)
		if err != nil {
			fmt.Printf("failed to add %s to roller.toml: %v", key, err)
			return err
		}
	}

	/* ------------------------------ Create Init Files ---------------------------- */
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

	archivePath, err := globalutils.ExpandHomePath(path)
	if err != nil {
		return "", err
	}

	if _, err := os.Stat(archivePath); os.IsNotExist(err) {
		return "", err
	}

	return archivePath, nil
}
