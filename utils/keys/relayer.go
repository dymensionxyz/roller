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
		"test",
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
	lookFor := fmt.Sprintf("no keys found for chain %s", chainId)

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
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.RollappRelayer,
			ChainBinary: rollappConfig.RollappBinary,
			Type:        rollappConfig.RollappVMType,
		},
		consts.KeysIds.HubRelayer: {
			Dir:         path.Join(rollappConfig.Home, consts.ConfigDirName.Relayer),
			ID:          consts.KeysIds.HubRelayer,
			ChainBinary: consts.Executables.Dymension,
			Type:        consts.SDK_ROLLAPP,
		},
	}
}

func GetRelayerKeys(rollappConfig roller.RollappConfig) ([]KeyInfo, error) {
	relayerAddresses := make([]KeyInfo, 0)
	relayerKeys := GetRelayerKeysConfig(rollappConfig)

	relayerHubAddress, err := GetRelayerAddressInfo(
		relayerKeys[consts.KeysIds.HubRelayer],
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

func GenerateRelayerKeys(rollappConfig roller.RollappConfig) ([]KeyInfo, error) {
	pterm.Info.Println("creating relayer keys")
	var relayerAddresses []KeyInfo
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
