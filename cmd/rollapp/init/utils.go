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
	celestialightclient "github.com/dymensionxyz/roller/data_layer/celestia/lightclient"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

func runInit(
	cmd *cobra.Command,
	env string,
	raResp rollapp.ShowRollappResponse,
) error {
	raID := raResp.Rollapp.RollappId

	home, err := filesystem.ExpandHomePath(cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String())
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

	// TODO: extract into util

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
	var initConfigPtr *roller.RollappConfig

	if env == consts.MockHubName {
		initConfigPtr, err = roller.GetMockRollappMetadata(
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
		initConfigPtr, err = rollapp.GetRollappMetadataFromChain(
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
	var addresses []keys.KeyInfo

	useExistingSequencerWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"would you like to import an existing sequencer key?",
	).Show()

	if useExistingSequencerWallet {
		kc, err := keys.NewKeyConfig(
			consts.ConfigDirName.HubKeys,
			consts.KeysIds.HubSequencer,
			consts.Executables.Dymension,
			consts.SDK_ROLLAPP,
			keys.WithRecover(),
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
		addresses, err = keys.GenerateMockSequencerKeys(initConfig)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return err
		}
	} else {
		if !useExistingSequencerWallet {
			addresses, err = keys.GenerateSequencersKeys(initConfig)
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
		"rollapp_id":      raID,
		"rollapp_binary":  strings.ToLower(consts.Executables.RollappEVM),
		"rollapp_vm_type": string(initConfigPtr.RollappVMType),
		"home":            home,

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

	PrintInitOutput(initConfig, addresses, initConfig.RollappID)

	return nil
}

func PrintInitOutput(
	rollappConfig roller.RollappConfig,
	addresses []keys.KeyInfo,
	rollappId string,
) {
	fmt.Printf(
		"💈 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n",
		rollappId,
	)

	if rollappConfig.HubData.ID == consts.MockHubID {
		roller.PrintTokenSupplyLine(rollappConfig)
		fmt.Println()
	}
	keys.PrintAddressesWithTitle(addresses)

	if rollappConfig.HubData.ID != consts.MockHubID {
		pterm.DefaultSection.WithIndentCharacter("🔔").
			Println("Please fund the addresses below to register and run the rollapp.")
		fa := initconfig.FormatAddresses(rollappConfig, addresses)
		for _, v := range fa {
			v.Print(keys.WithName())
		}
	}
}
