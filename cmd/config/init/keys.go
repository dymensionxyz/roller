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

func generateKeys(rollappId string, hubId string, rollappKeyPrefix string, excludeKeys ...string) (map[string]string, error) {
	keys := getDefaultKeysConfig(rollappId, hubId, rollappKeyPrefix)
	excludeKeysMap := make(map[string]struct{})
	for _, key := range excludeKeys {
		excludeKeysMap[key] = struct{}{}
	}
	addresses := make(map[string]string)
	for _, key := range keys {
		if _, exists := excludeKeysMap[key.keyId]; !exists {
			keyInfo, err := createKey(rollappId, key.dir, key.keyId, key.coinType)
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

type keyConfig struct {
	dir      string
	keyId    string
	coinType uint32
	prefix   string
}

func createKey(rollappId string, relativePath string, keyId string, coinType ...uint32) (keyring.Info, error) {
	var coinTypeVal = cosmosDefaultCointype
	if len(coinType) != 0 {
		coinTypeVal = coinType[0]
	}
	kr, err := keyring.New(
		rollappId,
		keyring.BackendTest,
		filepath.Join(getRollerRootDir(), relativePath),
		nil,
	)
	if err != nil {
		return nil, err
	}
	bip44Params := hd.NewFundraiserParams(0, coinTypeVal, 0)
	info, _, err := kr.NewMnemonic(keyId, keyring.English, bip44Params.String(), "", hd.Secp256k1)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func getDefaultKeysConfig(rollappId string, hubId string, rollappKeyPrefix string) []keyConfig {
	return []keyConfig{
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
			prefix:   rollappKeyPrefix,
		},
		{
			dir:      path.Join(configDirName.Relayer, relayerKeysDirName, rollappId),
			keyId:    keyNames.HubRelayer,
			coinType: cosmosDefaultCointype,
			prefix:   keyPrefixes.Hub,
		},
		{
			dir:      path.Join(configDirName.Relayer, relayerKeysDirName, hubId),
			keyId:    keyNames.RollappRelayer,
			coinType: evmCoinType,
			prefix:   rollappKeyPrefix,
		}, {

			dir:      path.Join(configDirName.DALightNode, relayerKeysDirName),
			keyId:    keyNames.DALightNode,
			coinType: cosmosDefaultCointype,
			prefix:   keyPrefixes.DA,
		},
	}
}
