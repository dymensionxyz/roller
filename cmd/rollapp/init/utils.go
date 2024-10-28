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
) error {
	raID := raResp.Rollapp.RollappId

	home, err := filesystem.ExpandHomePath(cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String())
	if err != nil {
		pterm.Error.Println("failed to expand home directory")
		return err
	}

	var hd consts.HubData
	// todo: refactor
	if env != "custom" {
		hd = consts.Hubs[env]
	} else {
		hd = customHubData
	}
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
		initConfigPtr, err = rollapp.PopulateRollerConfigWithRaMetadataFromChain(
			home,
			raID,
			hd,
		)
		if err != nil {
			errorhandling.PrettifyErrorIfExists(err)
			return err
		}
	}
	initConfig := *initConfigPtr

	/* ------------------------------ Generate keys ----------------------------- */
	var addresses []keys.KeyInfo
	var k []keys.KeyInfo

	addresses = append(addresses, k...)
	sequencerKeys, err := keys.GenerateSequencerKeys(home, env, initConfig)
	if err != nil {
		return err
	}
	addresses = append(addresses, sequencerKeys...)

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

	daData, err := datalayer.GetDaInfo(env, daBackend)
	if err != nil {
		return err
	}

	err = roller.PopulateConfig(home, raID, hd, *daData, string(initConfig.RollappVMType))
	if err != nil {
		return err
	}

	err = initConfig.ValidateConfig()
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

	if daKeyInfo != nil {
		addresses = append(addresses, *daKeyInfo)
	}

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
