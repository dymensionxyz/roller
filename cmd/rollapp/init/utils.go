package initrollapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	cmdutils "github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

// nolint: gocyclo
func runInit(cmd *cobra.Command, env string, raResp rollapp.ShowRollappResponse) error {
	raID := raResp.Rollapp.RollappId

	home, err := filesystem.ExpandHomePath(cmd.Flag(cmdutils.FlagNames.Home).Value.String())
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

	// TODO: extract into util
	isRootExist, err := filesystem.DirNotEmpty(home)
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return err
	}

	if isRootExist {
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

			err = servicemanager.RemoveServiceFiles(consts.RollappSystemdServices)
			if err != nil {
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
	// TODO: refactor
	var initConfigPtr *config.RollappConfig

	if env == consts.MockHubName {
		initConfigPtr, err = tomlconfig.GetMockRollappMetadata(
			home,
			raID,
			&hd,
			raResp.Rollapp.VmType,
		)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return err
		}
	} else {
		initConfigPtr, err = tomlconfig.GetRollappMetadataFromChain(
			home,
			raID,
			&hd,
		)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return err
		}
	}
	initConfig := *initConfigPtr

	/* ------------------------------ Generate keys ----------------------------- */
	var addresses []cmdutils.KeyInfo

	useExistingSequencerWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"would you like to import an existing sequencer key?",
	).Show()

	if useExistingSequencerWallet {
		kc, err := utils.NewKeyConfig(
			consts.ConfigDirName.HubKeys,
			consts.KeysIds.HubSequencer,
			consts.Executables.Dymension,
			consts.SDK_ROLLAPP,
			utils.WithRecover(),
		)
		if err != nil {
			return err
		}

		ki, err := keys.CreateAddressBinary(*kc, home)
		if err != nil {
			return err
		}

		ki.Print(cmdutils.WithName())
	}

	if initConfig.HubData.ID == "mock" {
		addresses, err = initconfig.GenerateMockSequencerKeys(initConfig)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return err
		}
	} else {
		if !useExistingSequencerWallet {
			addresses, err = initconfig.GenerateSequencersKeys(initConfig)
			if err != nil {
				return err
			}
		}
	}

	/* --------------------------- Initialize Rollapp -------------------------- */
	raSpinner, err := pterm.DefaultSpinner.Start("initializing rollapp client")
	if err != nil {
		return err
	}

	err = initconfig.InitializeRollappConfig(&initConfig, raResp)
	if err != nil {
		raSpinner.Fail("failed to initialize rollapp client")
		return err
	}

	// adds the sequencer address to the whitelists
	if env == "mock" {
		err = genesisutils.InitializeRollappGenesis(initConfig)
		if err != nil {
			return err
		}
	}

	// Initialize roller config
	as, err := genesisutils.GetGenesisAppState(home)
	if err != nil {
		return err
	}
	daBackend := as.RollappParams.Params.Da
	pterm.Info.Println("DA backend: ", daBackend)

	var daData consts.DaData
	var daNetwork string
	switch env {
	case "playground":
		if daBackend == string(consts.Celestia) {
			daNetwork = consts.DefaultCelestiaNetwork
		} else {
			return fmt.Errorf("unsupported DA backend: %s", daBackend)
		}
	case "mock":
		daNetwork = "mock"
	default:
		return fmt.Errorf("unsupported environment: %s", env)
	}

	daData = consts.DaNetworks[daNetwork]
	rollerTomlData := map[string]string{
		"rollapp_id":     raID,
		"rollapp_binary": strings.ToLower(consts.Executables.RollappEVM),
		"execution":      string(initConfigPtr.VMType),
		"home":           home,

		"HubData.id":              hd.ID,
		"HubData.api_url":         hd.API_URL,
		"HubData.rpc_url":         hd.RPC_URL,
		"HubData.archive_rpc_url": hd.ARCHIVE_RPC_URL,
		"HubData.gas_price":       hd.GAS_PRICE,

		"DA.backend":    string(daData.Backend),
		"DA.id":         string(daData.ID),
		"DA.api_url":    daData.ApiUrl,
		"DA.rpc_url":    daData.RpcUrl,
		"DA.state_node": daData.StateNode,
		"DA.gas_price":  daData.GasPrice,
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

	errorhandling.RunOnInterrupt(outputHandler.StopSpinner)
	err = initConfig.Validate()
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return err
	}

	/* ------------------------------ Initialize Local Hub ---------------------------- */
	// TODO: local hub is out of scope, implement as the last step
	// hub := cmd.Flag(FlagNames.HubID).Value.String()
	// if hub == consts.LocalHubName {
	// 	err := initLocalHub(initConfig)
	// 	utils.PrettifyErrorIfExists(err)
	// }

	raSpinner.Success("rollapp initialized successfully")

	/* ------------------------ Initialize DA light node ------------------------ */
	if env != "mock" {
		daSpinner, _ := pterm.DefaultSpinner.Start("initializing da light client")

		damanager := datalayer.NewDAManager(initConfig.DA.Backend, initConfig.Home)
		mnemonic, err := damanager.InitializeLightNodeConfig()
		if err != nil {
			return err
		}

		sequencers, err := sequencer.RegisteredRollappSequencersOnHub(raID, hd)
		if err != nil {
			return err
		}

		latestHeight, latestBlockIdHash, err := GetLatestDABlock(initConfig)
		if err != nil {
			return err
		}
		heightInt, err := strconv.Atoi(latestHeight)
		if err != nil {
			return err
		}

		celestiaConfigFilePath := filepath.Join(
			home,
			consts.ConfigDirName.DALightNode,
			"config.toml",
		)
		if len(sequencers.Sequencers) == 0 {
			pterm.Info.Println("no sequencers registered for the rollapp")
			pterm.Info.Println("using latest height for da-light-client configuration")

			pterm.Info.Printf(
				"da light client will be initialized at height %s, block hash %s",
				latestHeight,
				latestBlockIdHash,
			)

			err = UpdateCelestiaConfig(celestiaConfigFilePath, latestBlockIdHash, heightInt)
			if err != nil {
				return err
			}
		} else {
			daSpinner.UpdateText("checking for state update ")
			cmd := exec.Command(
				consts.Executables.Dymension,
				"q",
				"rollapp",
				"state",
				raID,
				"--index",
				"1",
				"--node",
				hd.RPC_URL,
				"--chain-id",
				hd.ID,
			)

			out, err := bash.ExecCommandWithStdout(cmd)
			if err != nil {
				if strings.Contains(out.String(), "NotFound") {
					pterm.Info.Printf(
						"no state found for %s, da light client will be initialized with latest height\n",
						raID,
					)
					err = UpdateCelestiaConfig(celestiaConfigFilePath, latestBlockIdHash, heightInt)
					if err != nil {
						return err
					}
				} else {
					return err
				}
			} else {
				daSpinner.UpdateText("state update found, extracting da height")

				var result Result
				if err := yaml.Unmarshal(out.Bytes(), &result); err != nil {
					pterm.Error.Println("failed to extract state update: ", err)
					return err
				}

				h, err := ExtractHeightfromDAPath(result.StateInfo.DAPath)
				if err != nil {
					pterm.Error.Println("failed to extract height from state update da path: ", err)
					return err
				}

				height, hash, err := GetDABlockByHeight(h, initConfig)
				if err != nil {
					pterm.Error.Println("failed to retrieve DA height: ", err)
					return err
				}

				heightInt, err := strconv.Atoi(height)
				if err != nil {
					return err
				}

				pterm.Info.Printf("the first %s state update has DA height of %s with hash %s\n", raID, height, hash)
				pterm.Info.Printf("updating %s \n", celestiaConfigFilePath)
				err = UpdateCelestiaConfig(celestiaConfigFilePath, hash, heightInt)
				if err != nil {
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
		daSpinner.Success("successfully initialized da light client")
	}
	/* ------------------------------ Print output ------------------------------ */

	outputHandler.PrintInitOutput(initConfig, addresses, initConfig.RollappID)

	return nil
}

func UpdateCelestiaConfig(file, hash string, height int) error {
	// Read existing config
	data, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	// Parse TOML into a map
	var config map[string]interface{}
	if err := toml.Unmarshal(data, &config); err != nil {
		return err
	}

	// Update DASer.SampleFrom
	if daser, ok := config["DASer"].(map[string]interface{}); ok {
		daser["SampleFrom"] = height
	}

	// Update Header.TrustedHash
	if header, ok := config["Header"].(map[string]interface{}); ok {
		header["TrustedHash"] = hash
	}

	// Write updated config
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	// nolint:errcheck
	defer f.Close()

	encoder := toml.NewEncoder(f)
	encoder.SetIndentTables(true)

	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}

// GetLatestDABlock returns the latest DA (Data Availability) block information.
// It executes the CelestiaApp command "q block --node" to retrieve the block data.
// It then extracts the block height and block ID hash from the JSON response.
// Returns the block height, block ID hash, and any error encountered during the process.
func GetLatestDABlock(raCfg config.RollappConfig) (string, string, error) {
	cmd := exec.Command(
		consts.Executables.CelestiaApp,
		"q", "block", "--node", raCfg.DA.RpcUrl, "--chain-id", string(raCfg.DA.ID),
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", "", err
	}

	var tx map[string]interface{}
	err = json.Unmarshal(out.Bytes(), &tx)
	if err != nil {
		return "", "", err
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
		return "", "", err
	}

	return height, blockIdHash, nil
}

// GetDABlockByHeight returns the DA (Data Availability) block information for the given height.
// It executes the CelestiaApp command "q block <height> --node" to retrieve the block data,
// where <height> is the input parameter.
// It then extracts the block height and block ID hash from the JSON response.
// Returns the block height, block ID hash, and any error encountered during the process.
func GetDABlockByHeight(h string, raCfg config.RollappConfig) (string, string, error) {
	cmd := exec.Command(
		consts.Executables.CelestiaApp,
		"q", "block", h, "--node", raCfg.DA.RpcUrl, "--chain-id", string(raCfg.DA.ID),
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", "", err
	}

	var tx map[string]interface{}
	err = json.Unmarshal(out.Bytes(), &tx)
	if err != nil {
		return "", "", err
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
		return "", "", err
	}

	return height, blockIdHash, nil
}

type BD struct {
	Height    string `yaml:"height"`
	StateRoot string `yaml:"stateRoot"`
}

type StateInfo struct {
	BDs struct {
		BD []BD `yaml:"BD"`
	} `yaml:"BDs"`
	DAPath         string `yaml:"DAPath"`
	CreationHeight string `yaml:"creationHeight"`
	NumBlocks      string `yaml:"numBlocks"`
	Sequencer      string `yaml:"sequencer"`
	StartHeight    string `yaml:"startHeight"`
	StateInfoIndex struct {
		Index     string `yaml:"index"`
		RollappId string `yaml:"rollappId"`
	} `yaml:"stateInfoIndex"`
	Status string `yaml:"status"`
}

type Result struct {
	StateInfo StateInfo `yaml:"stateInfo"`
}

func ExtractHeightfromDAPath(input string) (string, error) {
	parts := strings.Split(input, "|")
	if len(parts) < 2 {
		return "", fmt.Errorf("input string does not have enough parts")
	}
	return parts[1], nil
}
