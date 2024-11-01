package keys

import (
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
	fmt.Println(showKeyCommand.String())

	output, err := bash.ExecCommandWithStdout(showKeyCommand)
	if err != nil {
		return nil, err
	}

	return ParseAddressFromOutput(output)
}

func IsRlyAddressWithNameInKeyring(
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

	fmt.Println("out", out.String())
	lookFor := fmt.Sprintf("no keys found for chain %s", chainId)

	if out.String() == "" {
		return false, nil
	}

	return !strings.Contains(out.String(), lookFor), nil
}

// TODO: remove this struct as it's redundant to KeyInfo
type SecretAddressData struct {
	AddressData
	Mnemonic string
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
	relayerAddresses := make([]KeyInfo, 0)
	relayerKeys := GetRelayerKeysConfig(rollappConfig)

	relayerHubAddress, err := GetRelayerAddressInfo(
		relayerKeys[consts.KeysIds.HubRelayer],
		rollappConfig.HubData.ID,
	)
	if err != nil {
		return err
	}

	relayerAddresses = append(
		relayerAddresses, *relayerHubAddress,
	)

	pterm.Info.Println(
		"please fund the hub relayer key with at least 20 dym tokens: ",
	)
	for _, k := range relayerAddresses {
		k.Print(WithName())
	}

	return nil
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

func GenerateRelayerKeys(rollerData roller.RollappConfig) error {
	pterm.Info.Println("creating relayer keys")
	var createdRlyKeys []KeyInfo
	keys := GetRelayerKeysConfig(rollerData)

	for k, v := range keys {
		pterm.Info.Printf("checking %s in %s\n", k, v.Dir)

		switch v.ID {
		case consts.KeysIds.RollappRelayer:
			chainId := rollerData.RollappID
			isPresent, err := IsRlyAddressWithNameInKeyring(v, chainId)
			if err != nil {
				pterm.Error.Printf("failed to check address: %v\n", err)
				return err
			}

			if !isPresent {
				pterm.Info.Printf("creating %s in %s\n", k, v.Dir)
				key, err := AddRlyKey(v, rollerData.RollappID)
				if err != nil {
					pterm.Error.Printf("failed to add key: %v\n", err)
				}
				createdRlyKeys = append(createdRlyKeys, *key)
			}
		case consts.KeysIds.HubRelayer:
			chainId := rollerData.HubData.ID
			isPresent, err := IsRlyAddressWithNameInKeyring(v, chainId)
			if err != nil {
				pterm.Error.Printf("failed to check address: %v\n", err)
				return err
			}

			if !isPresent {
				pterm.Info.Printf("creating %s in %s\n", k, v.Dir)
				key, err := AddRlyKey(v, rollerData.HubData.ID)
				if err != nil {
					pterm.Error.Printf("failed to add key: %v\n", err)
				}
				createdRlyKeys = append(createdRlyKeys, *key)
			}
		default:
			return fmt.Errorf("invalid key name: %s", v.ID)
		}
	}

	if len(createdRlyKeys) != 0 {
		for _, key := range createdRlyKeys {
			key.Print(WithMnemonic(), WithName())
		}
	}

	return nil
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
