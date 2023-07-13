package initconfig

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"os/exec"
	"path"
	"path/filepath"
)

func generateKeys(rollappConfig config.RollappConfig) ([]utils.AddressData, error) {
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

func generateSequencersKeys(initConfig config.RollappConfig) ([]utils.AddressData, error) {
	keys := getSequencerKeysConfig(initConfig)
	addresses := make([]utils.AddressData, 0)
	for _, key := range keys {
		var address string
		var err error
		address, err = createAddressBinary(key, initConfig.Home)
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

func getSequencerKeysConfig(rollappConfig config.RollappConfig) []utils.KeyConfig {
	return []utils.KeyConfig{
		{
			Dir:         consts.ConfigDirName.HubKeys,
			ID:          consts.KeysIds.HubSequencer,
			ChainBinary: consts.Executables.Dymension,
		},
		{
			Dir:         consts.ConfigDirName.Rollapp,
			ID:          consts.KeysIds.RollappSequencer,
			ChainBinary: rollappConfig.RollappBinary,
		},
	}
}

func getRelayerKeysConfig(rollappConfig config.RollappConfig) map[string]utils.KeyConfig {
	return map[string]utils.KeyConfig{
		consts.KeysIds.RollappRelayer: {
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.RollappRelayer,
			ChainBinary: rollappConfig.RollappBinary,
		},
		consts.KeysIds.HubRelayer: {
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.HubRelayer,
			ChainBinary: consts.Executables.Dymension,
		},
	}
}

func createAddressBinary(keyConfig utils.KeyConfig, home string) (string, error) {
	args := []string{
		"keys", "add", keyConfig.ID, "--keyring-backend", "test",
		"--keyring-dir", filepath.Join(home, keyConfig.Dir),
		"--output", "json",
	}
	if keyConfig.ChainBinary == consts.Executables.Dymension {
		args = append(args, "--algo", consts.AlgoTypes.Secp256k1)
	}
	createKeyCommand := exec.Command(keyConfig.ChainBinary, args...)
	out, err := utils.ExecBashCommand(createKeyCommand)
	if err != nil {
		return "", err
	}
	return utils.ParseAddressFromOutput(out)
}

func generateRelayerKeys(rollappConfig config.RollappConfig) ([]utils.AddressData, error) {
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

func getAddRlyKeyCmd(keyConfig utils.KeyConfig, chainID string) *exec.Cmd {
	// TODO: Add support for custom EVM rollapp binaries (#196)
	var coinType = "118"
	if keyConfig.ChainBinary == consts.Executables.RollappEVM {
		coinType = "60"
	}
	return exec.Command(
		consts.Executables.Relayer,
		"keys",
		"add",
		chainID,
		keyConfig.ID,
		"--home",
		keyConfig.Dir,
		"--coin-type",
		coinType,
	)
}
