package initconfig

import (
	"os/exec"
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func generateKeys(initConfig utils.RollappConfig, excludeKeys ...string) (map[string]string, error) {
	keys := getDefaultKeysConfig(initConfig)
	excludeKeysMap := make(map[string]struct{})
	for _, key := range excludeKeys {
		excludeKeysMap[key] = struct{}{}
	}
	addresses := make(map[string]string)
	for _, key := range keys {
		if _, exists := excludeKeysMap[key.ID]; !exists {
			if key.Prefix == consts.AddressPrefixes.Rollapp {
				address, err := createAddressBinary(key, consts.Executables.RollappEVM, initConfig.Home)
				if err != nil {
					return nil, err
				}
				addresses[key.ID] = address
			} else {
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
	}
	//relayerAddresses, err := generateRelayerKeys(initConfig)
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

func getDefaultKeysConfig(initConfig utils.RollappConfig) []utils.KeyConfig {
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
			Dir:      path.Join(consts.ConfigDirName.Relayer, consts.KeysDirName, initConfig.HubData.ID),
			ID:       consts.KeyNames.HubRelayer,
			CoinType: consts.CoinTypes.Cosmos,
			Prefix:   consts.AddressPrefixes.Hub,
		},
		{
			Dir:      path.Join(consts.ConfigDirName.Relayer, consts.KeysDirName, initConfig.RollappID),
			ID:       consts.KeyNames.RollappRelayer,
			CoinType: consts.CoinTypes.EVM,
			Prefix:   consts.AddressPrefixes.Rollapp,
		},
	}
}

func createAddressBinary(keyConfig utils.KeyConfig, binaryPath string, home string) (string, error) {
	createKeyCommand := exec.Command(binaryPath, "keys", "add", keyConfig.ID, "--keyring-backend", "test",
		"--keyring-dir", filepath.Join(home, keyConfig.Dir), "--output", "json")
	out, err := utils.ExecBashCommand(createKeyCommand)
	if err != nil {
		return "", err
	}
	return utils.ParseAddressFromOutput(out)
}

//func generateRelayerKeys(rollappConfig utils.RollappConfig) (map[string]string, error) {
//
//}
//
//func getAddRlyKeyCmd(rollappConfig utils.RollappConfig) *exec.Cmd {
//	return exec.Command(
//		consts.Executables.Relayer,
//		consts.KeysDirName,
//		"add",
//		rollappConfig.RollappID,
//		"relayer-rollapp-keyy",
//		"--home",
//		"~/.roller/relayer",
//	)
//}
