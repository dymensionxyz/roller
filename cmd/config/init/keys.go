package initconfig

import (
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func generateKeys(initConfig InitConfig, excludeKeys ...string) (map[string]string, error) {
	keys := getDefaultKeysConfig(initConfig)
	excludeKeysMap := make(map[string]struct{})
	for _, key := range excludeKeys {
		excludeKeysMap[key] = struct{}{}
	}
	addresses := make(map[string]string)
	for _, key := range keys {
		if _, exists := excludeKeysMap[key.ID]; !exists {
			keyInfo, err := createKey(key, initConfig.Home)
			if err != nil {
				return nil, err
			}
			formattedAddress, err := utils.KeyInfoToBech32Address(keyInfo, key.Prefix)
			if err != nil {
				return nil, err
			}
			addresses[key.ID] = formattedAddress
		}
	}
	return addresses, nil
}

func createKey(keyConfig utils.KeyConfig, home string) (keyring.Info, error) {
	kr, err := keyring.New(
		"",
		keyring.BackendTest,
		filepath.Join(home, keyConfig.Dir),
		nil,
	)
	if err != nil {
		return nil, err
	}
	bip44Params := hd.NewFundraiserParams(0, keyConfig.CoinType, 0)
	info, _, err := kr.NewMnemonic(keyConfig.ID, keyring.English, bip44Params.String(), "", hd.Secp256k1)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func getDefaultKeysConfig(initConfig InitConfig) []utils.KeyConfig {
	return []utils.KeyConfig{
		{
			Dir:      consts.ConfigDirName.Rollapp,
			ID:       consts.KeyNames.HubSequencer,
			CoinType: consts.CoinTypes.Cosmos,
			Prefix:   consts.AddressPrefixes.Hub,
		},
		{
			Dir:      consts.ConfigDirName.Rollapp,
			ID:       consts.KeyNames.RollappSequencer,
			CoinType: consts.CoinTypes.EVM,
			Prefix:   consts.AddressPrefixes.Rollapp,
		},
		{
			Dir:      path.Join(consts.ConfigDirName.Relayer, KeysDirName, HubData.ID),
			ID:       consts.KeyNames.HubRelayer,
			CoinType: consts.CoinTypes.Cosmos,
			Prefix:   consts.AddressPrefixes.Hub,
		},
		{
			Dir:      path.Join(consts.ConfigDirName.Relayer, KeysDirName, initConfig.RollappID),
			ID:       consts.KeyNames.RollappRelayer,
			CoinType: consts.CoinTypes.EVM,
			Prefix:   consts.AddressPrefixes.Rollapp,
		}, {

			Dir:      path.Join(consts.ConfigDirName.DALightNode, KeysDirName),
			ID:       consts.KeyNames.DALightNode,
			CoinType: consts.CoinTypes.Cosmos,
			Prefix:   consts.AddressPrefixes.DA,
		},
	}
}

func initializeKeys(initConfig InitConfig) map[string]string {
	if initConfig.CreateDALightNode {
		addresses, err := generateKeys(initConfig)
		if err != nil {
			panic(err)
		}
		return addresses
	} else {
		addresses, err := generateKeys(initConfig, consts.KeyNames.DALightNode)
		if err != nil {
			panic(err)
		}
		return addresses
	}
}
