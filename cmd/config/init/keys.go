package init

import (
	"os"
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func generateKeys(rollappId string, hubId string, excludeKeys ...string) error {
	keys := getDefaultKeys(rollappId, hubId)
	excludeKeysMap := make(map[string]struct{})
	for _, key := range excludeKeys {
		excludeKeysMap[key] = struct{}{}
	}

	for _, key := range keys {
		if _, exists := excludeKeysMap[key.keyId]; !exists {
			_, err := createKey(rollappId, key.dir, key.keyId, key.coinType)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

type keyConfig struct {
	dir      string
	keyId    string
	coinType uint32
}

func getDALightNodeKey() keyConfig {
	return keyConfig{}
}

func createKey(rollappId string, relativePath string, keyId string, coinType ...uint32) (keyring.Info, error) {
	var coinTypeVal = cosmosDefaultCointype
	if len(coinType) != 0 {
		coinTypeVal = coinType[0]
	}
	kr, err := keyring.New(
		rollappId,
		keyring.BackendTest,
		filepath.Join(os.Getenv("HOME"), relativePath),
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

func getDefaultKeys(rollappId string, hubId string) []keyConfig {
	return []keyConfig{
		{
			dir:      configDirName.Rollapp,
			keyId:    keyNames.HubSequencer,
			coinType: cosmosDefaultCointype,
		},
		{
			dir:      configDirName.Rollapp,
			keyId:    keyNames.RollappSequencer,
			coinType: evmCoinType,
		},
		{
			dir:      path.Join(configDirName.Relayer, relayerKeysDirName, rollappId),
			keyId:    "relayer-hub-key",
			coinType: cosmosDefaultCointype,
		},
		{
			dir:      path.Join(configDirName.Relayer, relayerKeysDirName, hubId),
			keyId:    keyNames.RollappRelayer,
			coinType: evmCoinType,
		}, {

			dir:      path.Join(configDirName.DALightNode, relayerKeysDirName),
			keyId:    keyNames.lightNode,
			coinType: cosmosDefaultCointype,
		},
	}
}
