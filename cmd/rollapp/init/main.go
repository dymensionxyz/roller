package initrollapp

import (
	"bufio"
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

var Cmd = &cobra.Command{
	Use:   "init",
	Short: "Inititlize RollApp locally",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		err := initconfig.AddFlags(cmd)
		if err != nil {
			fmt.Println("failed to add flags")
			return
		}
		reader := bufio.NewReader(os.Stdin)

		fmt.Println("Do you already have rollapp config? (y/n)")
		resp, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			return
		}

		resp = strings.TrimSpace(resp)
		resp = strings.ToLower(resp)

		if resp == "n" || resp == "no" {
			fmt.Println(
				`To generate a RollApp configuration file go to <website>
or run 'rollapp config' to expose the UI on localhost:11133.
after configuration files are generated, rerun the 'init' command`,
			)
			return
		}

		if resp == "y" || resp == "yes" {
			fmt.Println(
				"provide a path to the configuration archive file downloaded from <website>",
			)
			fp, err := reader.ReadString('\n')
			if err != nil {
				return
			}

			fp = strings.TrimSpace(fp)
			if fp == "" {
				fmt.Println("no path was provided")
				return
			}

			archivePath, err := expandHomePath(fp)
			if err != nil {
				return
			}

			if _, err := os.Stat(archivePath); os.IsNotExist(err) {
				fmt.Printf("the file %s does not exist. \n", fp)
				return
			}

			err = runInit(cmd, args, archivePath)
			if err != nil {
				fmt.Printf("failed to initialize the RollApp: %v\n", err)
				return
			}

			return
		}

		fmt.Println("invalid or no input")
	},
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

// in runInit I parse the entire genesis creator zip file twice to extract
// the file this looks awful but since the archive has only 2 files it's
// kinda fine
func runInit(cmd *cobra.Command, args []string, configArchivePath string) error {
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
		return err
	}

	initConfigPtr, err := initconfig.GetInitConfig(cmd, args)
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

	err = archives.ExtractFileFromNestedTar(
		configArchivePath,
		"genesis.json",
		filepath.Join(home, consts.ConfigDirName.Rollapp, "config"),
	)
	if err != nil {
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
