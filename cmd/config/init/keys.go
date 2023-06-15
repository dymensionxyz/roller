package initconfig

import (
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func generateKeys(rollappConfig utils.RollappConfig) (map[string]string, error) {
	sequencerAddresses, err := generateSequencersKeys(rollappConfig)
	if err != nil {
		return nil, err
	}
	relayerAddresses, err := generateRelayerKeys(rollappConfig)
	if err != nil {
		return nil, err
	}
	return utils.MergeMaps(sequencerAddresses, relayerAddresses), nil
}

func generateSequencersKeys(initConfig utils.RollappConfig) (map[string]string, error) {
	keys := getSequencerKeysConfig()
	addresses := make(map[string]string)
	for _, key := range keys {
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

func getSequencerKeysConfig() []utils.KeyConfig {
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
	}
}

func getRelayerKeysConfig(rollappConfig utils.RollappConfig) map[string]utils.KeyConfig {
	return map[string]utils.KeyConfig{
		consts.KeyNames.RollappRelayer: {
			Dir:      path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:       consts.KeyNames.RollappRelayer,
			CoinType: consts.CoinTypes.EVM,
			Prefix:   consts.AddressPrefixes.Rollapp,
		},
		consts.KeyNames.HubRelayer: {
			Dir:      path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:       consts.KeyNames.HubRelayer,
			CoinType: consts.CoinTypes.Cosmos,
			Prefix:   consts.AddressPrefixes.Hub,
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

func generateRelayerKeys(rollappConfig utils.RollappConfig) (map[string]string, error) {
	relayerAddresses := make(map[string]string)
	keys := getRelayerKeysConfig(rollappConfig)
	createRollappKeyCmd := getAddRlyKeyCmd(keys[consts.KeyNames.RollappRelayer], rollappConfig.RollappID)
	createHubKeyCmd := getAddRlyKeyCmd(keys[consts.KeyNames.HubRelayer], rollappConfig.HubData.ID)
	out, err := utils.ExecBashCommand(createRollappKeyCmd)
	if err != nil {
		return nil, err
	}
	relayerRollappAddress, err := utils.ParseAddressFromOutput(out)
	if err != nil {
		return nil, err
	}
	relayerAddresses[consts.KeyNames.RollappRelayer] = relayerRollappAddress
	out, err = utils.ExecBashCommand(createHubKeyCmd)
	if err != nil {
		return nil, err
	}
	relayerHubAddress, err := utils.ParseAddressFromOutput(out)
	if err != nil {
		return nil, err
	}
	relayerAddresses[consts.KeyNames.HubRelayer] = relayerHubAddress
	return relayerAddresses, err
}

func getAddRlyKeyCmd(keyConfig utils.KeyConfig, chainID string) *exec.Cmd {
	return exec.Command(
		consts.Executables.Relayer,
		consts.KeysDirName,
		"add",
		chainID,
		keyConfig.ID,
		"--home",
		keyConfig.Dir,
		"--coin-type",
		strconv.Itoa(int(keyConfig.CoinType)),
	)
}
