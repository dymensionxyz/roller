package keys

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"
)

func createSequencersKeys(rollerData roller.RollappConfig) ([]KeyInfo, error) {
	sequencerKeys := GetSequencerKeysConfig()
	addresses := make([]KeyInfo, 0)
	for _, key := range sequencerKeys {
		var address *KeyInfo
		var err error
		address, err = CreateAddressBinary(key, rollerData.Home)
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

func GenerateSequencerKeys(home, env string, rollerData roller.RollappConfig) ([]KeyInfo, error) {
	var k []KeyInfo
	var err error

	if env == "mock" {
		k, err = generateMockSequencerKeys(rollerData)
		if err != nil {
			return nil, err
		}
	} else {
		k, err = generateRaSequencerKeys(home, rollerData)
		if err != nil {
			return nil, err
		}
	}

	return k, nil
}

func generateMockSequencerKeys(initConfig roller.RollappConfig) ([]KeyInfo, error) {
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

func generateRaSequencerKeys(home string, rollerData roller.RollappConfig) ([]KeyInfo, error) {
	useExistingSequencerWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"would you like to import an existing sequencer key?",
	).Show()

	var addr []KeyInfo
	var err error

	if useExistingSequencerWallet {
		kc, err := NewKeyConfig(
			consts.ConfigDirName.HubKeys,
			consts.KeysIds.HubSequencer,
			consts.Executables.Dymension,
			consts.SDK_ROLLAPP,
			WithRecover(),
		)
		if err != nil {
			return nil, err
		}

		ki, err := kc.Create(home)
		if err != nil {
			return nil, err
		}
		addr = append(addr, *ki)
	} else {
		addr, err = createSequencersKeys(rollerData)
		if err != nil {
			return nil, err
		}
	}

	return addr, nil
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
