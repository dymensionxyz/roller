package init

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
			dir:      configDirName.Rollapp,
			keyId:    keyNames.HubSequencer,
			coinType: cosmosDefaultCointype,
			prefix:   keyPrefixes.Hub,
		},
		{
			dir:      configDirName.Rollapp,
			keyId:    keyNames.RollappSequencer,
			coinType: evmCoinType,
			prefix:   initConfig.RollappPrefix,
		},
		{
			dir:      path.Join(configDirName.Relayer, relayerKeysDirName, initConfig.RollappID),
			keyId:    keyNames.HubRelayer,
			coinType: cosmosDefaultCointype,
			prefix:   keyPrefixes.Hub,
		},
		{
			dir:      path.Join(configDirName.Relayer, relayerKeysDirName, initConfig.HubID),
			keyId:    keyNames.RollappRelayer,
			coinType: evmCoinType,
			prefix:   initConfig.RollappPrefix,
		}, {

			dir:      path.Join(configDirName.DALightNode, relayerKeysDirName),
			keyId:    keyNames.DALightNode,
			coinType: cosmosDefaultCointype,
			prefix:   keyPrefixes.DA,
		},
	}
}
