package relayer

import (
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

// FIXME: move to relayer package

// TODO: this should be a method on Relayer.Relayer in `relayer_manager`
func InitializeRelayer(home string, rollerData roller.RollappConfig) error {
	// at this point it is safe to assume that
	// relayer is being initialized on a sequencer node
	// there is an existing roller config that can be used as the data source
	relayerHome := filepath.Join(home, consts.ConfigDirName.Relayer)
	isRelayerInitialized, err := filesystem.DirNotEmpty(relayerHome)
	if err != nil {
		pterm.Error.Printf("failed to check %s: %v\n", relayerHome, err)
		return err
	}

	if isRelayerInitialized {
		return nil
	}

	pterm.Info.Println("initializing relayer config")
	err = relayer.InitializeConfig(
		relayer.ChainConfig{
			ID:            rollerData.RollappID,
			RPC:           consts.DefaultRollappRPC,
			Denom:         rollerData.BaseDenom,
			AddressPrefix: rollerData.Bech32Prefix,
			GasPrices:     "2000000000",
		}, relayer.ChainConfig{
			ID:            rollerData.HubData.ID,
			RPC:           rollerData.HubData.RpcUrl,
			Denom:         consts.Denoms.Hub,
			AddressPrefix: consts.AddressPrefixes.Hub,
			GasPrices:     rollerData.HubData.GasPrice,
		}, home,
	)
	if err != nil {
		pterm.Error.Printf(
			"failed to initialize relayer config: %v\n",
			err,
		)
		return err
	}

	return nil
}

func EnsureKeysArePresentAndFunded(
	rollerData roller.RollappConfig,
) (map[string]keys.KeyInfo, error) {
	ck, err := keys.GenerateRelayerKeys(rollerData)
	if err != nil {
		pterm.Error.Printf("failed to create relayer keys: %v\n", err)
		return nil, err
	}

	err = keys.GetRelayerKeysToFund(rollerData)
	if err != nil {
		pterm.Error.Printf("failed to retrieve relayer keys to fund: %v\n", err)
		return nil, err
	}

	proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
		WithDefaultText(
			"press 'y' when the wallets are funded",
		).Show()
	if !proceed {
		return nil, fmt.Errorf("cancelled by user")
	}

	err = relayer.VerifyRelayerBalances(rollerData.HubData)
	if err != nil {
		return nil, err
	}

	return ck, nil
}
