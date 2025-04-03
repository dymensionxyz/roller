package initrollapp

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

func runInit(
	home, env string,
	customHubData consts.HubData,
	raResp rollapp.ShowRollappResponse,
	kb consts.SupportedKeyringBackend,
) error {
	raID := raResp.Rollapp.RollappId

	var hd consts.HubData
	if env != "custom" {
		hd = consts.Hubs[env]
	} else {
		hd = customHubData
	}

	// TODO: should set keyring as well
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

	// save the genesis file in the rollapp directory
	if env != consts.MockHubName {
		err = genesisutils.DownloadGenesis(home, ic.GenesisUrl)
		if err != nil {
			return err
		}
	}

	/* ------------------------------ Generate keys ----------------------------- */
	var addresses []keys.KeyInfo

	sequencerKeys, err := initSequencerKeys(home, env, ic)
	if err != nil {
		return err
	}

	addresses = append(addresses, sequencerKeys...)

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
	err := InitializeRollappNode(initConfig)
	if err != nil {
		return fmt.Errorf("failed to initialize rollapp client: %v", err)
	}
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
