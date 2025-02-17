package initrollapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	celestialightclient "github.com/dymensionxyz/roller/data_layer/celestia/lightclient"
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
	customHubData consts.HubData,
	raResp rollapp.ShowRollappResponse,
	kb consts.SupportedKeyringBackend,
) error {
	raID := raResp.Rollapp.RollappId

	home, err := filesystem.ExpandHomePath(cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String())
	if err != nil {
		pterm.Error.Println("failed to expand home directory")
		return err
	}

	var hd consts.HubData
	if env != "custom" {
		hd = consts.Hubs[env]
	} else {
		hd = customHubData
	}

	ic, err := prepareConfig(env, home, raID, hd, raResp)
	if err != nil {
		return err
	}

	/* --------------------------- Initialize Rollapp -------------------------- */
	err = initRollapp(ic, raResp, env, home, raID, hd, kb)
	if err != nil {
		return err
	}

	/* ------------------------------ Generate keys ----------------------------- */
	var addresses []keys.KeyInfo

	sequencerKeys, err := initSequencerKeys(home, env, ic)
	if err != nil {
		return err
	}

	if env == "mock" {
		err = genesisutils.InitializeRollappGenesis(ic)
		if err != nil {
			return err
		}
	}

	addresses = append(addresses, sequencerKeys...)

	/* ------------------------------ Initialize Local Hub ---------------------------- */
	// TODO: local hub is out of scope, implement as the last step
	// hub := cmd.Flag(FlagNames.HubID).Value.String()
	// if hub == consts.LocalHubName {
	// 	err := initLocalHub(initConfig)
	// 	utils.PrettifyErrorIfExists(err)
	// }

	/* ------------------------ Initialize DA light node ------------------------ */
	daKeyInfo, err := celestialightclient.Initialize(env, ic)
	if err != nil {
		return err
	}

	if daKeyInfo != nil {
		addresses = append(addresses, *daKeyInfo)
	}

	/* ------------------------------ Print output ------------------------------ */
	PrintInitOutput(ic, addresses, ic.RollappID)

	return nil
}

func initSequencerKeys(home string, env string, ic roller.RollappConfig) ([]keys.KeyInfo, error) {
	err := keys.CreateSequencerOsKeyringPswFile(home)
	if err != nil {
		return nil, err
	}
	sequencerKeys, err := keys.GenerateSequencerKeys(home, env, ic)
	if err != nil {
		return nil, err
	}
	return sequencerKeys, nil
}

func initRollapp(
	initConfig roller.RollappConfig,
	raResp rollapp.ShowRollappResponse,
	env string,
	home string,
	raID string,
	hd consts.HubData,
	kb consts.SupportedKeyringBackend,
) error {
	raSpinner, err := pterm.DefaultSpinner.Start("initializing rollapp client")
	if err != nil {
		return err
	}

	err = initconfig.InitializeRollappConfig(&initConfig, raResp)
	if err != nil {
		raSpinner.Fail("failed to initialize rollapp client")
		return err
	}

	as, err := genesisutils.GetGenesisAppState(home)
	if err != nil {
		return err
	}

	daBackend := as.RollappParams.Params.Da
	pterm.Info.Println("DA backend: ", daBackend)

	daData, err := datalayer.GetDaInfo(env, daBackend)
	if err != nil {
		return err
	}

	err = roller.PopulateConfig(home, raID, hd, *daData, string(initConfig.RollappVMType), kb)
	if err != nil {
		return err
	}

	err = initConfig.ValidateConfig()
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return err
	}

	// nolint: errcheck
	raSpinner.Stop()
	pterm.DefaultSection.WithIndentCharacter("❗️").
		Println("Below is the validator key of this node. It should be backed up so the node can be recovered in case of failure.")
	jsonFilePath := filepath.Join(home, "rollapp", "config", "priv_validator_key.json")
	jsonData, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return err
	}

	var jsonContent map[string]interface{}
	err = json.Unmarshal(jsonData, &jsonContent)
	if err != nil {
		return err
	}

	jsonString, err := json.MarshalIndent(jsonContent, "", "  ")
	if err != nil {
		return err
	}

	fmt.Println(string(jsonString))

	isBackedUp, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"press 'y' when you have backed up the validator key",
	).Show()

	if !isBackedUp {
		return errors.New("cancelled by user")
	}

	raSpinner.Success("rollapp initialized successfully")
	return nil
}

func prepareConfig(
	env string,
	home string,
	raID string,
	hd consts.HubData,
	raResp rollapp.ShowRollappResponse,
) (roller.RollappConfig, error) {
	var initConfig roller.RollappConfig

	if env == consts.MockHubName {
		ic, err := roller.GetMockRollappMetadata(
			home,
			raID,
			&hd,
			raResp.Rollapp.VmType,
		)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return roller.RollappConfig{}, err
		}
		initConfig = *ic
	} else {
		ic, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
			home,
			raID,
			hd,
		)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return roller.RollappConfig{}, err
		}
		initConfig = *ic
	}
	return initConfig, nil
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
}
