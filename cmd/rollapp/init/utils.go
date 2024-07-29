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

type Options struct {
	configArchivePath string
	useMockSettlement bool
}

type Option func(*Options)

func WithConfig(configArchivePath string) Option {
	return func(o *Options) {
		o.configArchivePath = configArchivePath
	}
}

func WithMockSettlement() Option {
	return func(o *Options) {
		o.useMockSettlement = true
	}
}

func defaultOptions() Options {
	return Options{
		configArchivePath: "",
	}
}

// in runInit I parse the entire genesis creator zip file twice to extract
// the file this looks awful but since the archive has only 2 files it's
// kinda fine
func runInit(cmd *cobra.Command, opts ...Option) error {
	options := defaultOptions()
	for _, opt := range opts {
		opt(&options)
	}

	home := utils.GetRollerRootDir()
	outputHandler := initconfig.NewOutputHandler(false)
	configArchivePath := options.configArchivePath

	defer outputHandler.StopSpinner()

	// TODO: extract into util
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
			return err
		}
		if shouldOverwrite {
			err = os.RemoveAll(home)
			if err != nil {
				utils.PrettifyErrorIfExists(err)
				return err
			}
		} else {
			os.Exit(0)
		}
	}

	// nolint:gofumpt
	err = os.MkdirAll(home, 0o755)
	if err != nil {
		utils.PrettifyErrorIfExists(err)
		return err
	}

	if configArchivePath != "" {
		err = archives.ExtractFileFromNestedTar(
			configArchivePath,
			config.RollerConfigFileName,
			home,
		)
		if err != nil {
			return err
		}
	} else {
		mockRollerConfig := NewMockRollerConfig()
		err := WriteMockRollerconfigToFile(mockRollerConfig)
		if err != nil {
			return err
		}
	}

	initConfigPtr, err := initconfig.GetInitConfig(cmd, options.useMockSettlement)
	if err != nil {
		utils.PrettifyErrorIfExists(err)
		return err
	}

	initConfig := *initConfigPtr

	utils.RunOnInterrupt(outputHandler.StopSpinner)
	outputHandler.StartSpinner(consts.SpinnerMsgs.UniqueIdVerification)
	err = initConfig.Validate()
	if err != nil {
		utils.PrettifyErrorIfExists(err)
		return err
	}

	// TODO: create all dirs here
	outputHandler.StartSpinner(" Initializing RollApp configuration files...")

	/* ------------------------------ Generate keys ----------------------------- */
	addresses, err := initconfig.GenerateSequencersKeys(initConfig)
	if err != nil {
		utils.PrettifyErrorIfExists(err)
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

	if configArchivePath != "" {
		// genesis creator archive
		err = archives.ExtractFileFromNestedTar(
			configArchivePath,
			"genesis.json",
			filepath.Join(home, consts.ConfigDirName.Rollapp, "config"),
		)
		if err != nil {
			return err
		}
	}

	// adds the sequencer address to the whitelists
	err = initconfig.UpdateGenesisParams(home, &initConfig)
	if err != nil {
		return err
	}

	rollerConfigFilePath := filepath.Join(utils.GetRollerRootDir(), "roller.toml")
	err = global_utils.UpdateFieldInToml(rollerConfigFilePath, "home", utils.GetRollerRootDir())
	if err != nil {
		fmt.Println("failed to add home to roller.toml: ", err)
		return err
	}

	err = global_utils.UpdateFieldInToml(rollerConfigFilePath, "HubData.ID", initConfig.HubData.ID)
	if err != nil {
		fmt.Println("failed to add HubData.ID to roller.toml: ", err)
		return err
	}

	err = global_utils.UpdateFieldInToml(
		rollerConfigFilePath,
		"HubData.rpc_url",
		initConfig.HubData.RPC_URL,
	)
	if err != nil {
		fmt.Println("failed to add HubData.RpcUrl to roller.toml: ", err)
		return err
	}

	err = global_utils.UpdateFieldInToml(
		rollerConfigFilePath,
		"HubData.gas_price",
		initConfig.HubData.GAS_PRICE,
	)
	if err != nil {
		fmt.Println("failed to add HubData.GasPrices to roller.toml: ", err)
		return err
	}

	err = global_utils.UpdateFieldInToml(
		rollerConfigFilePath,
		"da",
		strings.ToLower(string(initConfig.DA)),
	)
	if err != nil {
		fmt.Println("failed to add HubData.RpcUrl to roller.toml: ", err)
		return err
	}

	err = global_utils.UpdateFieldInToml(
		rollerConfigFilePath,
		"rollapp_binary",
		strings.ToLower(consts.Executables.RollappEVM),
	)
	if err != nil {
		fmt.Println("failed to add HubData.RpcUrl to roller.toml: ", err)
		return err
	}

	/* ------------------------------ Create Init Files ---------------------------- */
	err = WriteDenomMetadata(initConfig)
	if err != nil {
		fmt.Println("failed to create denom metadata: ", err)
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
