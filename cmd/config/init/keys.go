package initconfig

import (
	"os/exec"
	"path"
	"path/filepath"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func generateKeys(rollappConfig utils.RollappConfig) ([]utils.AddressData, error) {
	sequencerAddresses, err := generateSequencersKeys(rollappConfig)
	if err != nil {
		return nil, err
	}
	relayerAddresses, err := generateRelayerKeys(rollappConfig)
	if err != nil {
		return nil, err
	}
	return append(sequencerAddresses, relayerAddresses...), nil
}

func generateSequencersKeys(initConfig utils.RollappConfig) ([]utils.AddressData, error) {
	keys := getSequencerKeysConfig()
	addresses := make([]utils.AddressData, 0)
	for _, key := range keys {
		var address string
		var err error
		if key.Prefix == consts.AddressPrefixes.Rollapp {
			address, err = createAddressBinary(key, consts.Executables.RollappEVM, initConfig.Home)
		} else {
			address, err = createAddressBinary(key, consts.Executables.Dymension, initConfig.Home)
		}
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, utils.AddressData{
			Addr: address,
			Name: key.ID,
		})
	}
	return addresses, nil
}

func getSequencerKeysConfig() []utils.CreateKeyConfig {
	return []utils.CreateKeyConfig{
		{
			Dir:      consts.ConfigDirName.HubKeys,
			ID:       consts.KeysIds.HubSequencer,
			CoinType: consts.CoinTypes.EVM,
			Algo:     consts.AlgoTypes.Secp256k1,
			Prefix:   consts.AddressPrefixes.Hub,
		},
		{
			Dir:      consts.ConfigDirName.Rollapp,
			ID:       consts.KeysIds.RollappSequencer,
			CoinType: consts.CoinTypes.EVM,
			Algo:     consts.AlgoTypes.Ethsecp256k1,
			Prefix:   consts.AddressPrefixes.Rollapp,
		},
	}
}

func getRelayerKeysConfig(rollappConfig utils.RollappConfig) map[string]utils.CreateKeyConfig {
	return map[string]utils.CreateKeyConfig{
		consts.KeysIds.RollappRelayer: {
			Dir:      path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:       consts.KeysIds.RollappRelayer,
			CoinType: consts.CoinTypes.EVM,
			Algo:     consts.AlgoTypes.Ethsecp256k1,
			Prefix:   consts.AddressPrefixes.Rollapp,
		},
		consts.KeysIds.HubRelayer: {
			Dir:      path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:       consts.KeysIds.HubRelayer,
			CoinType: consts.CoinTypes.Cosmos,
			Algo:     consts.AlgoTypes.Secp256k1,
			Prefix:   consts.AddressPrefixes.Hub,
		},
	}
}

func createAddressBinary(keyConfig utils.CreateKeyConfig, binaryPath string, home string) (string, error) {
	createKeyCommand := exec.Command(binaryPath, "keys", "add", keyConfig.ID, "--keyring-backend", "test",
		"--keyring-dir", filepath.Join(home, keyConfig.Dir), "--algo", keyConfig.Algo, "--output", "json")
	out, err := utils.ExecBashCommand(createKeyCommand)
	if err != nil {
		return "", err
	}
	return utils.ParseAddressFromOutput(out)
}

func generateRelayerKeys(rollappConfig utils.RollappConfig) ([]utils.AddressData, error) {
	relayerAddresses := make([]utils.AddressData, 0)
	keys := getRelayerKeysConfig(rollappConfig)
	createRollappKeyCmd := getAddRlyKeyCmd(keys[consts.KeysIds.RollappRelayer], rollappConfig.RollappID)
	createHubKeyCmd := getAddRlyKeyCmd(keys[consts.KeysIds.HubRelayer], rollappConfig.HubData.ID)
	out, err := utils.ExecBashCommand(createRollappKeyCmd)
	if err != nil {
		return nil, err
	}
	relayerRollappAddress, err := utils.ParseAddressFromOutput(out)
	if err != nil {
		return nil, err
	}
	relayerAddresses = append(relayerAddresses, utils.AddressData{
		Addr: relayerRollappAddress,
		Name: consts.KeysIds.RollappRelayer,
	})
	out, err = utils.ExecBashCommand(createHubKeyCmd)
	if err != nil {
		return nil, err
	}
	relayerHubAddress, err := utils.ParseAddressFromOutput(out)
	if err != nil {
		return nil, err
	}
	relayerAddresses = append(relayerAddresses, utils.AddressData{
		Addr: relayerHubAddress,
		Name: consts.KeysIds.HubRelayer,
	})
	return relayerAddresses, err
}

func getAddRlyKeyCmd(keyConfig utils.CreateKeyConfig, chainID string) *exec.Cmd {
	return exec.Command(
		consts.Executables.Relayer,
		"keys",
		"add",
		chainID,
		keyConfig.ID,
		"--home",
		keyConfig.Dir,
		"--coin-type",
		strconv.Itoa(int(keyConfig.CoinType)),
	)
}
