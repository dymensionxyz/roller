package initrollapp

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	celestialightclient "github.com/dymensionxyz/roller/data_layer/celestia/lightclient"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

// FIXME: move to utils/rollapp

func runInit(
	cmd *cobra.Command,
	env string,
	customHubData consts.HubData,
	raResp rollapp.ShowRollappResponse,
	kb consts.SupportedKeyringBackend,
) error {

	home, err := filesystem.ExpandHomePath(cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String())
	if err != nil {
		pterm.Error.Println("failed to expand home directory")
		return err
	}

	raID := raResp.Rollapp.RollappId

	var hd consts.HubData
	if env != "custom" {
		hd = consts.Hubs[env]
	} else {
		hd = customHubData
	}

	// FIXME: should get keyring as well
	ic, err := prepareConfig(env, home, raID, hd, raResp)
	if err != nil {
		return err
	}

	err = ic.ValidateConfig()
	if err != nil {
		return err
	}

	err = roller.WriteConfigToDisk(ic)
	if err != nil {
		return err
	}

	/* --------------------------- Initialize Rollapp -------------------------- */
	err = initRollapp(ic)
	if err != nil {
		return err
	}

	/* ------------------------------ Generate keys ----------------------------- */
	var addresses []keys.KeyInfo

	sequencerKeys, err := initSequencerKeys(home, env, ic)
	if err != nil {
		return err
	}

	addresses = append(addresses, sequencerKeys...)

	/* ------------------------ Initialize DA light node ------------------------ */

	// Generalize DA initialization logic
	switch ic.DA.Backend {
	case consts.Celestia:
		// Initialize Celestia light client
		daKeyInfo, err := celestialightclient.Initialize(env, ic)
		if err != nil {
			return fmt.Errorf("failed to initialize Celestia light client: %w", err)
		}

		// Append DA account address if available
		if daKeyInfo != nil {
			addresses = append(addresses, *daKeyInfo)
		}

	case consts.Avail:
		// Initialize DAManager for Avail
		damanager := datalayer.NewDAManager(consts.Avail, home, kb)

		// Retrieve DA account address
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return fmt.Errorf("failed to get Avail account address: %w", err)
		}

		// Append DA account address if available
		if daAddress != nil {
			addresses = append(addresses, keys.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			})
		}

	default:
		return fmt.Errorf("unsupported DA backend: %s", ic.DA.Backend)
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
) error {
	raSpinner, err := pterm.DefaultSpinner.Start("initializing rollapp client")
	if err != nil {
		return err
	}

	err = InitializeRollappNode(initConfig)
	if err != nil {
		raSpinner.Fail("failed to initialize rollapp client")
		return err
	}

	pterm.Info.Println(fmt.Sprintf("Config: %v", initConfig))
	fmt.Printf("config: %v\n", initConfig)

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
	var (
		ic  *roller.RollappConfig
		err error
	)

	if env == consts.MockHubName {
		ic, err = roller.GetMockRollappMetadata(
			home,
			raID,
			&hd,
			raResp.Rollapp.VmType,
		)
	} else {
		ic, err = rollapp.PopulateRollerConfigWithRaMetadataFromChain(
			home,
			raID,
			hd,
		)
	}
	if err != nil {
		errorhandling.PrettifyErrorIfExists(err)
		return roller.RollappConfig{}, err
	}
	return *ic, nil
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
