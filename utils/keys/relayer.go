package keys

import (
	"errors"
	"fmt"
	"os/exec"
	"path"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/roller"
)

func GetRelayerAddressInfo(keyConfig KeyConfig, chainId string) (*KeyInfo, error) {
	showKeyCommand := exec.Command(
		keyConfig.ChainBinary,
		"keys",
		"show",
		keyConfig.ID,
		"--keyring-backend",
		string(keyConfig.KeyringBackend),
		"--keyring-dir",
		filepath.Join(keyConfig.Dir, "keys", chainId),
		"--output",
		"json",
	)

	output, err := bash.ExecCommandWithStdout(showKeyCommand)
	if err != nil {
		return nil, err
	}

	return ParseAddressFromOutput(output)
}

func IsRlyAddressWithNameInKeyring(
	keyName string,
	info KeyConfig,
	chainId string,
) (bool, error) {
	cmd := exec.Command(
		consts.Executables.Relayer,
		"keys", "list", chainId, "--home", info.Dir,
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return false, err
	}

	if out.String() == "" {
		return false, nil
	}

	return strings.Contains(out.String(), keyName), nil
}

// TODO: remove this struct as it's redundant to KeyInfo
type SecretAddressData struct {
	AddressData
	Mnemonic string
}

func GetHubRelayerAddress(home string, kc KeyConfig) (string, error) {
	rlyAddr, err := kc.Address(home)
	if err != nil {
		return "", err
	}

	return rlyAddr, nil
}

func GetRelayerData(home string, kc KeyConfig, hd consts.HubData) ([]AccountData, error) {
	rlyAddr, err := GetHubRelayerAddress(home, kc)
	if err != nil {
		return nil, err
	}

	relayerBalance, err := QueryBalance(
		ChainQueryConfig{
			Binary: consts.Executables.Dymension,
			Denom:  consts.Denoms.Hub,
			RPC:    hd.RpcUrl,
		}, rlyAddr,
	)
	if err != nil {
		return nil, err
	}
	return []AccountData{
		{
			Address: rlyAddr,
			Balance: *relayerBalance,
		},
	}, nil
}

func GetRelayerKeysConfig(rollappConfig roller.RollappConfig) map[string]KeyConfig {
	return map[string]KeyConfig{
		consts.KeysIds.RollappRelayer: {
			Dir:            path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:             consts.KeysIds.RollappRelayer,
			ChainBinary:    rollappConfig.RollappBinary,
			Type:           rollappConfig.RollappVMType,
			KeyringBackend: consts.SupportedKeyringBackends.Test,
		},
		consts.KeysIds.HubRelayer: {
			Dir:            path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:             consts.KeysIds.HubRelayer,
			ChainBinary:    consts.Executables.Dymension,
			Type:           consts.SDK_ROLLAPP,
			KeyringBackend: consts.SupportedKeyringBackends.Test,
		},
	}
}

func GetRelayerKeysToFund(rollappConfig roller.RollappConfig) error {
	relayerKeys := GetRelayerKeysConfig(rollappConfig)

	rhki, err := GetRelayerAddressInfo(
		relayerKeys[consts.KeysIds.HubRelayer],
		rollappConfig.HubData.ID,
	)
	if err != nil {
		return err
	}

	for {
		cqc := ChainQueryConfig{
			Binary: consts.Executables.Dymension,
			Denom:  consts.Denoms.Hub,
			RPC:    rollappConfig.HubData.RpcUrl,
		}
		balance, err := QueryBalance(cqc, rhki.Address)
		if err != nil {
			return err
		}

		pterm.Info.Printf(
			"current balance: %s\nnecessary balance: >0\n",
			balance.String(),
		)

		if !balance.Amount.IsPositive() {
			pterm.Info.Println(
				"please fund the addresses below to operate the relayer.",
			)
			rhki.Print(WithName())
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallets are funded",
				).Show()
			if !proceed {
				return errors.New("cancelled by user")
			}
		} else {
			break
		}
	}

	return nil
}

func GenerateRelayerKeys(rollerData roller.RollappConfig) (map[string]KeyInfo, error) {
	pterm.Info.Println("creating relayer keys")
	createdRlyKeys := map[string]KeyInfo{}
	keys := GetRelayerKeysConfig(rollerData)

	for k, v := range keys {
		switch v.ID {
		case consts.KeysIds.RollappRelayer:
			chainId := rollerData.RollappID

			useExistingWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
				"would you like to import an existing relayer key for Rollapp?",
			).Show()

			if useExistingWallet {
				mnemonic, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
					"> Enter your bip39 mnemonic",
				).Show()
				ki, err := restoreRelayerKeyIfNotPresent(k, chainId, mnemonic, v)
				if err != nil {
					return nil, err
				}
				createdRlyKeys[consts.KeysIds.RollappRelayer] = *ki
			} else {
				ki, err := createRelayerKeyIfNotPresent(k, chainId, v)
				if err != nil {
					return nil, err
				}
				createdRlyKeys[consts.KeysIds.RollappRelayer] = *ki
			}
		case consts.KeysIds.HubRelayer:
			chainId := rollerData.HubData.ID
			useExistingWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
				"would you like to import an existing relayer key for Hub?",
			).Show()

			if useExistingWallet {
				mnemonic, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
					"> Enter your bip39 mnemonic",
				).Show()
				ki, err := restoreRelayerKeyIfNotPresent(k, chainId, mnemonic, v)
				if err != nil {
					return nil, err
				}
				createdRlyKeys[consts.KeysIds.RollappRelayer] = *ki
			} else {
				ki, err := createRelayerKeyIfNotPresent(k, chainId, v)
				if err != nil {
					return nil, err
				}
				createdRlyKeys[consts.KeysIds.RollappRelayer] = *ki
			}
		default:
			return nil, fmt.Errorf("invalid key name: %s", v.ID)
		}
	}

	if len(createdRlyKeys) != 0 {
		for _, key := range createdRlyKeys {
			key.Print(WithMnemonic(), WithName())
		}
	}

	return createdRlyKeys, nil
}

func createRelayerKeyIfNotPresent(
	keyName, chainID string,
	kc KeyConfig,
) (*KeyInfo, error) {
	isPresent, err := IsRlyAddressWithNameInKeyring(keyName, kc, chainID)
	var ki KeyInfo
	if err != nil {
		pterm.Error.Printf("failed to check address: %v\n", err)
		return nil, err
	}

	if !isPresent {
		key, err := AddRlyKey(kc, chainID)
		if err != nil {
			pterm.Error.Printf("failed to add key: %v\n", err)
		}

		ki = *key
	} else {
		key, err := GetRelayerAddressInfo(
			kc,
			chainID,
		)
		if err != nil {
			return nil, err
		}

		ki = *key
	}
	return &ki, nil
}

func restoreRelayerKeyIfNotPresent(
	keyName, chainID, mnemonic string,
	kc KeyConfig,
) (*KeyInfo, error) {
	isPresent, err := IsRlyAddressWithNameInKeyring(keyName, kc, chainID)
	var ki KeyInfo
	if err != nil {
		pterm.Error.Printf("failed to check address: %v\n", err)
		return nil, err
	}

	if !isPresent {
		key, err := RestoreRlyKey(kc, chainID, mnemonic)
		if err != nil {
			pterm.Error.Printf("failed to restore key: %v\n", err)
		}

		ki = *key
	} else {
		key, err := GetRelayerAddressInfo(
			kc,
			chainID,
		)
		if err != nil {
			return nil, err
		}

		ki = *key
	}
	return &ki, nil
}

func getAddRlyKeyCmd(keyConfig KeyConfig, chainID string) *exec.Cmd {
	coinType := "60"
	if keyConfig.Type == consts.WASM_ROLLAPP {
		coinType = "118"
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

func AddRlyKey(kc KeyConfig, chainID string) (*KeyInfo, error) {
	addKeyCmd := getAddRlyKeyCmd(
		kc,
		chainID,
	)

	out, err := bash.ExecCommandWithStdout(addKeyCmd)
	if err != nil {
		return nil, err
	}

	ki, err := ParseAddressFromOutput(out)
	if err != nil {
		return nil, err
	}
	ki.Name = kc.ID

	return ki, nil
}

func getRestoreRlyKeyCmd(keyConfig KeyConfig, chainID, mnemonic string) *exec.Cmd {
	coinType := "60"
	if keyConfig.Type == consts.WASM_ROLLAPP {
		coinType = "118"
	}
	return exec.Command(
		consts.Executables.Relayer,
		"keys",
		"restore",
		chainID,
		keyConfig.ID,
		mnemonic,
		"--home",
		keyConfig.Dir,
		"--coin-type",
		coinType,
	)
}

func RestoreRlyKey(kc KeyConfig, chainID, mnemonic string) (*KeyInfo, error) {
	restoreKeyCmd := getRestoreRlyKeyCmd(
		kc,
		chainID,
		mnemonic,
	)

	out, err := bash.ExecCommandWithStdout(restoreKeyCmd)
	if err != nil {
		return nil, err
	}

	ki, err := ParseAddressFromOutput(out)
	if err != nil {
		return nil, err
	}
	ki.Name = kc.ID

	return ki, nil
}
