package keys

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
)

func GenerateSequencersKeys(initConfig roller.RollappConfig) ([]KeyInfo, error) {
	sequencerKeys := GetSequencerKeysConfig()
	addresses := make([]KeyInfo, 0)
	for _, key := range sequencerKeys {
		var address *KeyInfo
		var err error
		address, err = CreateAddressBinary(key, initConfig.Home)
		if err != nil {
			return nil, err
		}
		addresses = append(
			addresses, KeyInfo{
				Address:  address.Address,
				Name:     key.ID,
				Mnemonic: address.Mnemonic,
			},
		)
	}
	return addresses, nil
}

func GenerateMockSequencerKeys(initConfig roller.RollappConfig) ([]KeyInfo, error) {
	sequencerKeys := GetMockSequencerKeyConfig(initConfig)
	addresses := make([]KeyInfo, 0)
	for _, key := range sequencerKeys {
		var address *KeyInfo
		var err error
		address, err = CreateAddressBinary(key, initConfig.Home)
		if err != nil {
			return nil, err
		}
		addresses = append(
			addresses, KeyInfo{
				Address:  address.Address,
				Name:     key.ID,
				Mnemonic: address.Mnemonic,
			},
		)
	}
	return addresses, nil
}

func GetSequencerKeysConfig() []KeyConfig {
	return []KeyConfig{
		{
			Dir:         consts.ConfigDirName.HubKeys,
			ID:          consts.KeysIds.HubSequencer,
			ChainBinary: consts.Executables.Dymension,
			// Eventhough the hub can get evm signatures, we still use the native
			Type: consts.SDK_ROLLAPP,
		},
	}
}

func GetMockSequencerKeyConfig(rollappConfig roller.RollappConfig) []KeyConfig {
	return []KeyConfig{
		{
			Dir:         consts.ConfigDirName.Rollapp,
			ID:          consts.KeysIds.RollappSequencer,
			ChainBinary: rollappConfig.RollappBinary,
			Type:        rollappConfig.RollappVMType,
		},
	}
}

func GetSequencerPubKey(rollappConfig roller.RollappConfig) (string, error) {
	cmd := exec.Command(
		rollappConfig.RollappBinary,
		"dymint",
		"show-sequencer",
		"--home",
		filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
	)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(strings.ReplaceAll(string(out), "\n", ""), "\\", ""), nil
}
