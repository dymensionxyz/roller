package initrollapp

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	cmdutils "github.com/dymensionxyz/roller/cmd/utils"
	celestialightclient "github.com/dymensionxyz/roller/data_layer/celestia/lightclient"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/rollapp"
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

			err = filesystem.RemoveServiceFiles(consts.RollappSystemdServices)
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

		ki, err := kc.Create(home)
		if err != nil {
			return err
		}

		addresses = append(addresses, *ki)
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
			daNetwork = string(consts.CelestiaTestnet)
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
	daKeyInfo, err := celestialightclient.Initialize(env, initConfig)
	if err != nil {
		return err
	}

	addresses = append(addresses, *daKeyInfo)

	/* ------------------------------ Print output ------------------------------ */

	outputHandler.PrintInitOutput(initConfig, addresses, initConfig.RollappID)

	return nil
}
