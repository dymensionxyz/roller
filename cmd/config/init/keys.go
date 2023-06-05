package initconfig

import (
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func keyInfoToBech32Address(info keyring.Info, prefix string) (string, error) {
	pk := info.GetPubKey()
	bech32Address, err := bech32.ConvertAndEncode(prefix, pk.Bytes())
	if err != nil {
		return "", err
	}
	return bech32Address, nil
}

func generateKeys(initConfig InitConfig, excludeKeys ...string) (map[string]string, error) {
	keys := getDefaultKeysConfig(initConfig)
	excludeKeysMap := make(map[string]struct{})
	for _, key := range excludeKeys {
		excludeKeysMap[key] = struct{}{}
	}
	addresses := make(map[string]string)
	for _, key := range keys {
		if _, exists := excludeKeysMap[key.keyId]; !exists {
			keyInfo, err := createKey(key, initConfig.Home)
			if err != nil {
				return nil, err
			}
			formattedAddress, err := keyInfoToBech32Address(keyInfo, key.prefix)
			if err != nil {
				return nil, err
			}
			addresses[key.keyId] = formattedAddress
		}
	}
	return addresses, nil
}

type KeyConfig struct {
	dir      string
	keyId    string
	coinType uint32
	prefix   string
}

func createKey(keyConfig KeyConfig, home string) (keyring.Info, error) {
	kr, err := keyring.New(
		"",
		keyring.BackendTest,
		filepath.Join(home, keyConfig.dir),
		nil,
	)
	if err != nil {
		return nil, err
	}
	bip44Params := hd.NewFundraiserParams(0, keyConfig.coinType, 0)
	info, _, err := kr.NewMnemonic(keyConfig.keyId, keyring.English, bip44Params.String(), "", hd.Secp256k1)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func getDefaultKeysConfig(initConfig InitConfig) []KeyConfig {
	return []KeyConfig{
		{
			dir:      ConfigDirName.Rollapp,
			keyId:    KeyNames.HubSequencer,
			coinType: cosmosDefaultCointype,
			prefix:   addressPrefixes.Hub,
		},
		{
			dir:      ConfigDirName.Rollapp,
			keyId:    KeyNames.RollappSequencer,
			coinType: evmCoinType,
			prefix:   addressPrefixes.Rollapp,
		},
		{
			dir:      path.Join(ConfigDirName.Relayer, KeysDirName, initConfig.HubID),
			keyId:    KeyNames.HubRelayer,
			coinType: cosmosDefaultCointype,
			prefix:   addressPrefixes.Hub,
		},
		{
			dir:      path.Join(ConfigDirName.Relayer, KeysDirName, initConfig.RollappID),
			keyId:    KeyNames.RollappRelayer,
			coinType: evmCoinType,
			prefix:   addressPrefixes.Rollapp,
		}, {

			dir:      path.Join(ConfigDirName.DALightNode, KeysDirName),
			keyId:    KeyNames.DALightNode,
			coinType: cosmosDefaultCointype,
			prefix:   addressPrefixes.DA,
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
		addresses, err := generateKeys(initConfig, KeyNames.DALightNode)
		if err != nil {
			panic(err)
		}
		return addresses
	}
}