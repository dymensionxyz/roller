package initrollapp

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	celestialightclient "github.com/dymensionxyz/roller/data_layer/celestia/lightclient"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

func runInit(
	home, env string,
	hubData consts.HubData,
	raResp rollapp.ShowRollappResponse,
	kb consts.SupportedKeyringBackend,
	shouldGenerateSequencerAddress bool,
) error {
	raID := raResp.Rollapp.RollappId

	// TODO: should set keyring as well
	ic, err := prepareConfig(env, home, raID, hubData, raResp)
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

	sequencerKeys, err := initSequencerKeys(home, env, ic, shouldGenerateSequencerAddress)
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

	case consts.LoadNetwork:
		// Initialize DAManager for LoadNetwork
		damanager := datalayer.NewDAManager(consts.LoadNetwork, home, kb)

		// Retrieve DA account address
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return fmt.Errorf("failed to get LoadNetwork account address: %w", err)
		}

		// Append DA account address if available
		if daAddress != nil {
			addresses = append(addresses, keys.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			})
		}

	case consts.Bnb:
		// Initialize DAManager for Bnb
		damanager := datalayer.NewDAManager(consts.Bnb, home, kb)

		// Retrieve DA account address
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return fmt.Errorf("failed to get Bnb account address: %w", err)
		}

		// Append DA account address if available
		if daAddress != nil {
			addresses = append(addresses, keys.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			})
		}
	case consts.Sui:
		// Initialize DAManager for Sui
		damanager := datalayer.NewDAManager(consts.Sui, home, kb)

		// Retrieve DA account address
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return fmt.Errorf("failed to get Bnb account address: %w", err)
		}

		// Append DA account address if available
		if daAddress != nil {
			addresses = append(addresses, keys.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			})
		}
	case consts.Aptos:
		// Initialize DAManager for Aptos
		damanager := datalayer.NewDAManager(consts.Aptos, home, kb)

		// Retrieve DA account address
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return fmt.Errorf("failed to get Aptos account address: %w", err)
		}

		// Append DA account address if available
		if daAddress != nil {
			addresses = append(addresses, keys.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			})
		}
	case consts.Walrus:
		// Initialize DAManager for Walrus
		damanager := datalayer.NewDAManager(consts.Walrus, home, kb)

		// Retrieve DA account address
		daAddress, err := damanager.GetDAAccountAddress()
		if err != nil {
			return fmt.Errorf("failed to get Walrus account address: %w", err)
		}

		// Append DA account address if available
		if daAddress != nil {
			addresses = append(addresses, keys.KeyInfo{
				Name:    damanager.GetKeyName(),
				Address: daAddress.Address,
			})
		}
	case consts.Mock:
	default:
		return fmt.Errorf("unsupported DA backend: %s", ic.DA.Backend)
	}

	/* ------------------------------ Print output ------------------------------ */
	PrintInitOutput(ic, addresses, ic.RollappID)

	return nil
}

func initSequencerKeys(
	home string,
	env string,
	ic roller.RollappConfig,
	shouldGenerateSequencerAddress bool,
) ([]keys.KeyInfo, error) {
	err := keys.CreateSequencerOsKeyringPswFile(home)
	if err != nil {
		return nil, err
	}
	sequencerKeys, err := keys.GenerateSequencerKeys(home, env, ic, shouldGenerateSequencerAddress)
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
