package initconfig

import (
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

func keyInfoToBech32Address(info keyring.Info, prefix string) (string, error) {
	pk := info.GetPubKey()
	addr := types.AccAddress(pk.Address())
	bech32Address, err := bech32.ConvertAndEncode(prefix, addr.Bytes())
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
		if _, exists := excludeKeysMap[key.ID]; !exists {
			keyInfo, err := createKey(key, initConfig.Home)
			if err != nil {
				return nil, err
			}
			formattedAddress, err := keyInfoToBech32Address(keyInfo, key.Prefix)
			if err != nil {
				return nil, err
			}
			addresses[key.ID] = formattedAddress
		}
	}
	return addresses, nil
}

type KeyConfig struct {
	Dir      string
	ID       string
	CoinType uint32
	Prefix   string
}

func createKey(keyConfig KeyConfig, home string) (keyring.Info, error) {
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

func getDefaultKeysConfig(initConfig InitConfig) []KeyConfig {
	return []KeyConfig{
		{
			Dir:      ConfigDirName.Rollapp,
			ID:       KeyNames.HubSequencer,
			CoinType: CoinTypes.Cosmos,
			Prefix:   AddressPrefixes.Hub,
		},
		{
			Dir:      ConfigDirName.Rollapp,
			ID:       KeyNames.RollappSequencer,
			CoinType: CoinTypes.EVM,
			Prefix:   AddressPrefixes.Rollapp,
		},
		{
			Dir:      path.Join(ConfigDirName.Relayer, KeysDirName, HubData.ID),
			ID:       KeyNames.HubRelayer,
			CoinType: CoinTypes.Cosmos,
			Prefix:   AddressPrefixes.Hub,
		},
		{
			Dir:      path.Join(ConfigDirName.Relayer, KeysDirName, initConfig.RollappID),
			ID:       KeyNames.RollappRelayer,
			CoinType: CoinTypes.EVM,
			Prefix:   AddressPrefixes.Rollapp,
		}, {

			Dir:      path.Join(ConfigDirName.DALightNode, KeysDirName),
			ID:       KeyNames.DALightNode,
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
		addresses, err := generateKeys(initConfig, KeyNames.DALightNode)
		if err != nil {
			panic(err)
		}
		return addresses
	}
}

func GetAddress(keyConfig KeyConfig) (string, error) {
	kr, err := keyring.New(
		"",
		keyring.BackendTest,
		keyConfig.Dir,
		nil,
	)
	if err != nil {
		return "", err
	}
	keyInfo, err := kr.Key(keyConfig.ID)
	if err != nil {
		return "", err
	}
	formattedAddress, err := keyInfoToBech32Address(keyInfo, keyConfig.Prefix)
	if err != nil {
		return "", err
	}

	return formattedAddress, nil
}
