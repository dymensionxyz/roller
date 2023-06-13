package initconfig

import (
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/cmd/consts"
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
			Dir:      ConfigDirName.Rollapp,
			ID:       consts.KeyNames.HubSequencer,
			CoinType: CoinTypes.Cosmos,
			Prefix:   AddressPrefixes.Hub,
		},
		{
			Dir:      ConfigDirName.Rollapp,
			ID:       consts.KeyNames.RollappSequencer,
			CoinType: CoinTypes.EVM,
			Prefix:   AddressPrefixes.Rollapp,
		},
		{
			Dir:      path.Join(ConfigDirName.Relayer, KeysDirName, HubData.ID),
			ID:       consts.KeyNames.HubRelayer,
			CoinType: CoinTypes.Cosmos,
			Prefix:   AddressPrefixes.Hub,
		},
		{
			Dir:      path.Join(ConfigDirName.Relayer, KeysDirName, initConfig.RollappID),
			ID:       consts.KeyNames.RollappRelayer,
			CoinType: CoinTypes.EVM,
			Prefix:   AddressPrefixes.Rollapp,
		}, {

			Dir:      path.Join(ConfigDirName.DALightNode, KeysDirName),
			ID:       consts.KeyNames.DALightNode,
			CoinType: CoinTypes.Cosmos,
			Prefix:   AddressPrefixes.DA,
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
