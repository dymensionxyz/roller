package initconfig

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

func GenerateKeys(rollappConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	// var addresses []utils.KeyInfo

	sequencerAddresses, err := generateSequencersKeys(rollappConfig)
	if err != nil {
		fmt.Println("failed to generate sequencerAddresses")
		return nil, err
	}

	// addresses = append(addresses, sequencerAddresses...)

	// relayerAddresses, err := generateRelayerKeys(rollappConfig)
	// if err != nil {
	// 	return nil, err
	// }
	// addresses = append(addresses, relayerAddresses...)

	return sequencerAddresses, nil
}

func generateSequencersKeys(initConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	keys := getSequencerKeysConfig(initConfig)
	addresses := make([]utils.KeyInfo, 0)
	for _, key := range keys {
		var address *utils.KeyInfo
		var err error
		address, err = CreateAddressBinary(key, initConfig.Home)
		if err != nil {
			return nil, err
		}
		addresses = append(addresses, utils.KeyInfo{
			Address:  address.Address,
			Name:     key.ID,
			Mnemonic: address.Mnemonic,
		})
	}
	return addresses, nil
}

func getSequencerKeysConfig(rollappConfig config.RollappConfig) []utils.KeyConfig {
	if rollappConfig.HubData.ID == consts.MockHubID {
		return []utils.KeyConfig{
			{
				Dir:         consts.ConfigDirName.Rollapp,
				ID:          consts.KeysIds.RollappSequencer,
				ChainBinary: rollappConfig.RollappBinary,
				Type:        rollappConfig.VMType,
			},
		}
	}
	return []utils.KeyConfig{
		{
			Dir:         consts.ConfigDirName.HubKeys,
			ID:          consts.KeysIds.HubSequencer,
			ChainBinary: consts.Executables.Dymension,
			// Eventhough the hub can get evm signatures, we still use the native
			Type: config.SDK_ROLLAPP,
		},
		{
			Dir:         consts.ConfigDirName.Rollapp,
			ID:          consts.KeysIds.RollappSequencer,
			ChainBinary: rollappConfig.RollappBinary,
			Type:        rollappConfig.VMType,
		},
	}
}

// func getRelayerKeysConfig(rollappConfig config.RollappConfig) map[string]utils.KeyConfig {
// 	return map[string]utils.KeyConfig{
// 		consts.KeysIds.RollappRelayer: {
// 			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
// 			ID:          consts.KeysIds.RollappRelayer,
// 			ChainBinary: rollappConfig.RollappBinary,
// 			Type:        rollappConfig.VMType,
// 		},
// 		consts.KeysIds.HubRelayer: {
// 			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
// 			ID:          consts.KeysIds.HubRelayer,
// 			ChainBinary: consts.Executables.Dymension,
// 			Type:        config.SDK_ROLLAPP,
// 		},
// 	}
// }

func CreateAddressBinary(
	keyConfig utils.KeyConfig,
	home string,
) (*utils.KeyInfo, error) {
	args := []string{
		"keys", "add", keyConfig.ID, "--keyring-backend", "test",
		"--keyring-dir", filepath.Join(home, keyConfig.Dir),
		"--output", "json",
	}
	createKeyCommand := exec.Command(keyConfig.ChainBinary, args...)
	out, err := utils.ExecBashCommandWithStdout(createKeyCommand)
	if err != nil {
		return nil, err
	}
	return utils.ParseAddressFromOutput(out)
}

// func generateRelayerKeys(rollappConfig config.RollappConfig) ([]utils.AddressData, error) {
// 	relayerAddresses := make([]utils.AddressData, 0)
// 	keys := getRelayerKeysConfig(rollappConfig)
// 	createRollappKeyCmd := getAddRlyKeyCmd(
// 		keys[consts.KeysIds.RollappRelayer],
// 		rollappConfig.RollappID,
// 	)
// 	createHubKeyCmd := getAddRlyKeyCmd(keys[consts.KeysIds.HubRelayer], rollappConfig.HubData.ID)
// 	out, err := utils.ExecBashCommandWithStdout(createRollappKeyCmd)
// 	if err != nil {
// 		return nil, err
// 	}
// 	relayerRollappAddress, err := utils.ParseAddressFromOutput(out)
// 	if err != nil {
// 		return nil, err
// 	}
// 	relayerAddresses = append(relayerAddresses, utils.AddressData{
// 		Addr: relayerRollappAddress,
// 		Name: consts.KeysIds.RollappRelayer,
// 	})
// 	out, err = utils.ExecBashCommandWithStdout(createHubKeyCmd)
// 	if err != nil {
// 		return nil, err
// 	}
// 	relayerHubAddress, err := utils.ParseAddressFromOutput(out)
// 	if err != nil {
// 		return nil, err
// 	}
// 	relayerAddresses = append(relayerAddresses, utils.AddressData{
// 		Addr: relayerHubAddress,
// 		Name: consts.KeysIds.HubRelayer,
// 	})
// 	return relayerAddresses, err
// }

// func getAddRlyKeyCmd(keyConfig utils.KeyConfig, chainID string) *exec.Cmd {
// 	coinType := "118"
// 	if keyConfig.Type == config.EVM_ROLLAPP {
// 		coinType = "60"
// 	}
// 	return exec.Command(
// 		consts.Executables.Relayer,
// 		"keys",
// 		"add",
// 		chainID,
// 		keyConfig.ID,
// 		"--home",
// 		keyConfig.Dir,
// 		"--coin-type",
// 		coinType,
// 	)
// }
