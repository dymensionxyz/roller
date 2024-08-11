package initconfig

import (
	"fmt"
	"os/exec"
	"path"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
)

func GenerateSequencersKeys(initConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	keys := getSequencerKeysConfig(initConfig)
	addresses := make([]utils.KeyInfo, 0)
	for _, key := range keys {
		var address *utils.KeyInfo
		var err error
		address, err = CreateAddressBinary(key, initConfig.Home)
		if err != nil {
			return nil, err
		}
		addresses = append(
			addresses, utils.KeyInfo{
				Address:  address.Address,
				Name:     key.ID,
				Mnemonic: address.Mnemonic,
			},
		)
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
			Type: consts.SDK_ROLLAPP,
		},
		{
			Dir:         consts.ConfigDirName.HubKeys,
			ID:          consts.KeysIds.HubGenesis,
			ChainBinary: consts.Executables.Dymension,
			// Eventhough the hub can get evm signatures, we still use the native
			Type: consts.SDK_ROLLAPP,
		},
		{
			Dir:         consts.ConfigDirName.Rollapp,
			ID:          consts.KeysIds.RollappSequencer,
			ChainBinary: rollappConfig.RollappBinary,
			Type:        rollappConfig.VMType,
		},
	}
}

func getRelayerKeysConfig(rollappConfig config.RollappConfig) map[string]utils.KeyConfig {
	return map[string]utils.KeyConfig{
		consts.KeysIds.RollappRelayer: {
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.RollappRelayer,
			ChainBinary: rollappConfig.RollappBinary,
			Type:        rollappConfig.VMType,
		},
		consts.KeysIds.HubRelayer: {
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.HubRelayer,
			ChainBinary: consts.Executables.Dymension,
			Type:        consts.SDK_ROLLAPP,
		},
	}
}

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
	out, err := bash.ExecCommandWithStdout(createKeyCommand)
	if err != nil {
		return nil, err
	}
	return utils.ParseAddressFromOutput(out)
}

func GetRelayerKeys(rollappConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	pterm.Info.Println("getting relayer keys")
	relayerAddresses := make([]utils.KeyInfo, 0)
	keys := getRelayerKeysConfig(rollappConfig)

	showRollappKeyCmd := getShowRlyKeyCmd(
		keys[consts.KeysIds.RollappRelayer],
		rollappConfig.RollappID,
	)
	showHubKeyCmd := getShowRlyKeyCmd(
		keys[consts.KeysIds.HubRelayer],
		rollappConfig.HubData.ID,
	)

	out, err := bash.ExecCommandWithStdout(showRollappKeyCmd)
	if err != nil {
		pterm.Error.Printf("failed to retrieve rollapp key: %v\n", err)
	}
	relayerRollappAddress, err := utils.ParseAddressFromOutput(out)
	if err != nil {
		pterm.Error.Printf("failed to extract rollapp key: %v\n", err)
	}
	fmt.Println(relayerRollappAddress)

	out, err = bash.ExecCommandWithStdout(showHubKeyCmd)
	if err != nil {
		pterm.Error.Printf("failed to retrieve hub key: %v\n", err)
	}
	relayerHubAddress, err := utils.ParseAddressFromOutput(out)
	if err != nil {
		pterm.Error.Printf("failed to extract hub key: %v\n", err)
	}
	fmt.Println(relayerHubAddress)

	relayerAddresses = append(
		relayerAddresses, *relayerRollappAddress,
	)
	relayerAddresses = append(
		relayerAddresses, *relayerHubAddress,
	)

	return relayerAddresses, nil
}

func GenerateRelayerKeys(rollappConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	pterm.Info.Println("creating relayer keys")
	relayerAddresses := make([]utils.KeyInfo, 0)
	keys := getRelayerKeysConfig(rollappConfig)

	createRollappKeyCmd := getAddRlyKeyCmd(
		keys[consts.KeysIds.RollappRelayer],
		rollappConfig.RollappID,
	)
	createHubKeyCmd := getAddRlyKeyCmd(keys[consts.KeysIds.HubRelayer], rollappConfig.HubData.ID)

	pterm.Info.Println("creating relayer rollapp key")
	out, err := bash.ExecCommandWithStdout(createRollappKeyCmd)
	if err != nil {
		return nil, err
	}
	relayerRollappAddress, err := utils.ParseAddressFromOutput(out)
	relayerRollappAddress.Name = consts.KeysIds.RollappRelayer
	if err != nil {
		return nil, err
	}
	relayerAddresses = append(
		relayerAddresses, *relayerRollappAddress,
	)

	pterm.Info.Println("creating relayer hub key")
	out, err = bash.ExecCommandWithStdout(createHubKeyCmd)
	if err != nil {
		return nil, err
	}
	relayerHubAddress, err := utils.ParseAddressFromOutput(out)
	relayerHubAddress.Name = consts.KeysIds.HubRelayer
	if err != nil {
		return nil, err
	}
	relayerAddresses = append(
		relayerAddresses, *relayerHubAddress,
	)

	return relayerAddresses, nil
}

func getAddRlyKeyCmd(keyConfig utils.KeyConfig, chainID string) *exec.Cmd {
	coinType := "118"
	if keyConfig.Type == consts.EVM_ROLLAPP {
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

func getShowRlyKeyCmd(keyConfig utils.KeyConfig, chainID string) *exec.Cmd {
	coinType := "118"
	if keyConfig.Type == consts.EVM_ROLLAPP {
		coinType = "60"
	}
	return exec.Command(
		consts.Executables.Relayer,
		"keys",
		"show",
		chainID,
		keyConfig.ID,
		"--home",
		keyConfig.Dir,
		"--coin-type",
		coinType,
	)
}
