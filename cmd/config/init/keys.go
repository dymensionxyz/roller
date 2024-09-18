package initconfig

import (
	"os/exec"
	"path"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	keys2 "github.com/dymensionxyz/roller/utils/keys"
)

func GenerateSequencersKeys(initConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	keys := getSequencerKeysConfig(initConfig)
	addresses := make([]utils.KeyInfo, 0)
	for _, key := range keys {
		var address *utils.KeyInfo
		var err error
		address, err = keys2.CreateAddressBinary(key, initConfig.Home)
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
		// {
		// 	Dir:         consts.ConfigDirName.HubKeys,
		// 	ID:          consts.KeysIds.HubGenesis,
		// 	ChainBinary: consts.Executables.Dymension,
		// 	// Eventhough the hub can get evm signatures, we still use the native
		// 	Type: consts.SDK_ROLLAPP,
		// },
	}
}

func GetRelayerKeysConfig(rollappConfig config.RollappConfig) map[string]utils.KeyConfig {
	return map[string]utils.KeyConfig{
		consts.KeysIds.RollappRelayer: {
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.RollappRelayer,
			ChainBinary: rollappConfig.RollappBinary,
			Type:        consts.EVM_ROLLAPP,
		},
		consts.KeysIds.HubRelayer: {
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.HubRelayer,
			ChainBinary: consts.Executables.Dymension,
			Type:        consts.SDK_ROLLAPP,
		},
	}
}

func GetRelayerKeys(rollappConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	relayerAddresses := make([]utils.KeyInfo, 0)
	keys := GetRelayerKeysConfig(rollappConfig)

	relayerHubAddress, err := utils.GetRelayerAddressInfo(
		keys[consts.KeysIds.HubRelayer],
		rollappConfig.HubData.ID,
	)
	if err != nil {
		return nil, err
	}

	relayerAddresses = append(
		relayerAddresses, *relayerHubAddress,
	)

	return relayerAddresses, nil
}

func AddRlyKey(kc utils.KeyConfig, chainID string) (*utils.KeyInfo, error) {
	addKeyCmd := getAddRlyKeyCmd(
		kc,
		chainID,
	)

	out, err := bash.ExecCommandWithStdout(addKeyCmd)
	if err != nil {
		return nil, err
	}

	ki, err := utils.ParseAddressFromOutput(out)
	if err != nil {
		return nil, err
	}
	ki.Name = kc.ID

	return ki, nil
}

func GenerateRelayerKeys(rollappConfig config.RollappConfig) ([]utils.KeyInfo, error) {
	pterm.Info.Println("creating relayer keys")
	var relayerAddresses []utils.KeyInfo
	keys := GetRelayerKeysConfig(rollappConfig)

	pterm.Info.Println("creating relayer rollapp key")
	relayerRollappAddress, err := AddRlyKey(
		keys[consts.KeysIds.RollappRelayer],
		rollappConfig.RollappID,
	)
	if err != nil {
		return nil, err
	}

	relayerAddresses = append(
		relayerAddresses, *relayerRollappAddress,
	)

	pterm.Info.Println("creating relayer hub key")
	relayerHubAddress, err := AddRlyKey(keys[consts.KeysIds.HubRelayer], rollappConfig.HubData.ID)
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
	return exec.Command(
		consts.Executables.Relayer,
		"keys",
		"show",
		chainID,
		keyConfig.ID,
		"--home",
		keyConfig.Dir,
	)
}
