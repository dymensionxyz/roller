package keys

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/roller"
)

func createSequencersKeys(rollerData roller.RollappConfig) ([]KeyInfo, error) {
	sequencerKeys := GetSequencerKeysConfig(rollerData.KeyringBackend)
	addresses := make([]KeyInfo, 0)

	for _, key := range sequencerKeys {
		var address *KeyInfo
		var err error
		address, err = key.Create(rollerData.Home)
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

func CreateSequencerOsKeyringPswFile(home string) error {
	raFp := filepath.Join(home, string(consts.OsKeyringPwdFileNames.RollApp))
	return config.WritePasswordToFile(raFp)
}

func CreateDaOsKeyringPswFile(home string) error {
	daFp := filepath.Join(home, string(consts.OsKeyringPwdFileNames.Da))

	return config.WritePasswordToFile(daFp)
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
		address, err = key.Create(initConfig.Home)
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
			rollerData.KeyringBackend,
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

func GetSequencerKeysConfig(kb consts.SupportedKeyringBackend) []KeyConfig {
	return []KeyConfig{
		{
			Dir:         consts.ConfigDirName.HubKeys,
			ID:          consts.KeysIds.HubSequencer,
			ChainBinary: consts.Executables.Dymension,
			// Eventhough the hub can get evm signatures, we still use the native
			Type:           consts.SDK_ROLLAPP,
			KeyringBackend: kb,
		},
	}
}

func GetMockSequencerKeyConfig(rollappConfig roller.RollappConfig) []KeyConfig {
	return []KeyConfig{
		{
			Dir:            consts.ConfigDirName.Rollapp,
			ID:             consts.KeysIds.RollappSequencer,
			ChainBinary:    rollappConfig.RollappBinary,
			Type:           rollappConfig.RollappVMType,
			KeyringBackend: consts.SupportedKeyringBackends.Test,
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
