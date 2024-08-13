package initrollapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	cmdutils "github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

func runInit(cmd *cobra.Command, env string, raID string) error {
	home, err := globalutils.ExpandHomePath(cmd.Flag(cmdutils.FlagNames.Home).Value.String())
	if err != nil {
		pterm.Error.Println("failed to expand home directory")
		return err
	}
	rollerConfigFilePath := filepath.Join(home, consts.RollerConfigFileName)

	err = os.MkdirAll(home, 0o755)
	if err != nil {
		pterm.Error.Println("failed to create roller home directory: ", err)
		return err
	}

	// Check if the file already exists

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

			_, err := os.Stat(rollerConfigFilePath)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					// The file does not exist, so create it
					_, err = os.Create(rollerConfigFilePath)
					if err != nil {
						pterm.Error.Println(
							fmt.Sprintf("failed to create file %s: ", rollerConfigFilePath),
							err,
						)
						return err
					}
				} else {
					pterm.Error.Println(
						fmt.Sprintf("failed to check if file %s exists: ", rollerConfigFilePath),
						err,
					)
					return err
				}
			}
		} else {
			os.Exit(0)
		}
	}

	_, err = os.Stat(rollerConfigFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			pterm.Info.Println("roller.toml not found, creating")
			_, err := os.Create(rollerConfigFilePath)
			if err != nil {
				pterm.Error.Printf(
					"failed to create %s: %v", rollerConfigFilePath, err,
				)
				return err
			}
		}
	}

	hd := consts.Hubs[env]
	rollerTomlData := map[string]string{
		"rollapp_id":     raID,
		"rollapp_binary": strings.ToLower(consts.Executables.RollappEVM),
		"home":           home,

		"HubData.id":              hd.ID,
		"HubData.api_url":         hd.API_URL,
		"HubData.rpc_url":         hd.RPC_URL,
		"HubData.archive_rpc_url": hd.ARCHIVE_RPC_URL,
		"HubData.gas_price":       hd.GAS_PRICE,

		// TODO: create a separate config section for DA, similar to HubData
		"da": string(consts.Celestia),
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

	initConfigPtr, err := tomlconfig.LoadRollappMetadataFromChain(
		home,
		raID,
		&hd,
	)
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

	sequencers, err := sequencer.GetRegisteredSequencers(raID, hd)
	if err != nil {
		return err
	}

	if len(sequencers.Sequencers) == 0 {
		pterm.Info.Println("no sequencers registered for the rollapp")
		pterm.Info.Println("using latest height for da-light-client configuration")
		cmd := exec.Command(
			consts.Executables.CelestiaApp,
			"q", "block", "--node", celestia.DefaultCelestiaRPC,
		)

		out, err := bash.ExecCommandWithStdout(cmd)
		if err != nil {
			return err
		}

		var tx map[string]interface{}
		err = json.Unmarshal(out.Bytes(), &tx)
		if err != nil {
			log.Fatalf("Failed to unmarshal JSON: %v", err)
		}

		// Access tx.Block.Header.Height
		var height string
		if block, ok := tx["block"].(map[string]interface{}); ok {
			if header, ok := block["header"].(map[string]interface{}); ok {
				if h, ok := header["height"].(string); ok {
					height = h
				}
			}
		}

		// Access tx.BlockId.Hash
		var blockIdHash string
		if blockId, ok := tx["block_id"].(map[string]interface{}); ok {
			if hash, ok := blockId["hash"].(string); ok {
				blockIdHash = hash
			}
		}
		err = json.Unmarshal(out.Bytes(), &tx)
		if err != nil {
			return err
		}

		pterm.Info.Printf(
			"da light client will be initialized at height %s, block hash %s",
			height,
			blockIdHash,
		)
		heightInt, err := strconv.Atoi(height)
		if err != nil {
			return err
		}

		daFields := map[string]interface{}{
			"DASer.SampleFrom":   heightInt,
			"Header.TrustedHash": blockIdHash,
		}

		celestiaConfigFilePath := filepath.Join(
			home,
			consts.ConfigDirName.DALightNode,
			"config.toml",
		)

		for key, value := range daFields {
			err = globalutils.UpdateFieldInToml(
				celestiaConfigFilePath,
				key,
				value,
			)
			if err != nil {
				fmt.Printf("failed to add %s to roller.toml: %v", key, err)
				return err
			}
		}
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

	outputHandler.StopSpinner()
	/* ------------------------------ Create Init Files ---------------------------- */
	// 20240607 genesis is generated using the genesis-creator
	// err = initializeRollappGenesis(initConfig)
	// if err != nil {
	// 	return err
	// }

	// TODO: review, roller config is generated using genesis-creator
	// some of the config values should be moved there
	// err = config.Write(initConfig)
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
	if env != "mock" {
		err = tomlconfig.DownloadGenesis(home, initConfig)
		if err != nil {
			pterm.Error.Println("failed to download genesis file: ", err)
			return err
		}

		isChecksumValid, err := tomlconfig.CompareGenesisChecksum(home, raID, hd)
		if !isChecksumValid {
			return err
		}
	}

	/* ------------------------------ Print output ------------------------------ */

	outputHandler.PrintInitOutput(initConfig, addresses, initConfig.RollappID)

	return nil
}
