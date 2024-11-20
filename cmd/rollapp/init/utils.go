package initrollapp

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
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

	addresses = append(addresses, sequencerKeys...)

	/* ------------------------------ Initialize Local Hub ---------------------------- */
	// TODO: local hub is out of scope, implement as the last step
	// hub := cmd.Flag(FlagNames.HubID).Value.String()
	// if hub == consts.LocalHubName {
	// 	err := initLocalHub(initConfig)
	// 	utils.PrettifyErrorIfExists(err)
	// }

	/* ------------------------ Initialize DA light node ------------------------ */
	// daKeyInfo, err := celestialightclient.Initialize(env, ic)
	// if err != nil {
	// 	return err
	// }

	// if daKeyInfo != nil {
	// 	addresses = append(addresses, *daKeyInfo)
	// }

	damanager := datalayer.NewDAManager(consts.Avail, home, kb)
	_, err = damanager.InitializeLightNodeConfig()
	if err != nil {
		return err
	}
	daAddress, err := damanager.GetDAAccountAddress()
	if err != nil {
		return err
	}

	if daAddress != nil {
		addresses = append(addresses, keys.KeyInfo{
			Name:    damanager.GetKeyName(),
			Address: daAddress.Address,
		})
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

	fmt.Println("RA RESPP.....", raResp)

	err = initconfig.InitializeRollappConfig(&initConfig, raResp)
	if err != nil {
		raSpinner.Fail("failed to initialize rollapp client")
		return err
	}

	if env == "mock" {
		err = genesisutils.InitializeRollappGenesis(initConfig)
		if err != nil {
			return err
		}
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

	if rollappConfig.HubData.ID != consts.MockHubID {
		pterm.DefaultSection.WithIndentCharacter("🔔").
			Println("Please fund the addresses below to register and run the rollapp.")
		fa := initconfig.FormatAddresses(rollappConfig, addresses)
		for _, v := range fa {
			v.Print(keys.WithName())
		}
	}
}
