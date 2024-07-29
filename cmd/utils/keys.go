package utils

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

// KeyConfig struct store information about a wallet
// Dir refers to the keyringDir where the key is created
type KeyConfig struct {
	Dir string
	// TODO: this is not descriptive, Name would be more expressive
	ID          string
	ChainBinary string
	Type        config.VMType
}

// TODO: KeyInfo and AddressData seem redundant, should be moved into
// location

func GetRelayerAddress(home string, chainID string) (string, error) {
	showKeyCmd := exec.Command(
		consts.Executables.Relayer,
		"keys",
		"show",
		chainID,
		"--home",
		filepath.Join(home, consts.ConfigDirName.Relayer),
	)
	out, err := ExecBashCommandWithStdout(showKeyCmd)
	if err != nil {
		pterm.Error.Printf("no relayer address found: %v", err)
		return "", err
	}
	return strings.TrimSuffix(out.String(), "\n"), err
}

type AddressData struct {
	Name string
	Addr string
}

func GetSequencerPubKey(rollappConfig config.RollappConfig) (string, error) {
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

func GetAddressPrefix(binaryPath string) (string, error) {
	cmd := exec.Command(binaryPath, "debug", "addr", "ffffffffffffff")
	out, err := ExecBashCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}
	lines := strings.Split(out.String(), "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "Bech32 Acc:") {
			prefix := strings.Split(strings.TrimSpace(strings.Split(line, ":")[1]), "1")[0]
			return strings.TrimSpace(prefix), nil
		}
	}
	return "", fmt.Errorf("could not find address prefix in binary debug command output")
}

func GetExportKeyCmdBinary(keyID, keyringDir, binary string) *exec.Cmd {
	flags := getExportKeyFlags(keyringDir)
	var commandStr string
	if binary == consts.Executables.CelKey {
		commandStr = fmt.Sprintf("%s export %s %s", binary, keyID, flags)
	} else {
		commandStr = fmt.Sprintf("%s keys export %s %s", binary, keyID, flags)
	}
	cmdStr := fmt.Sprintf("yes | %s", commandStr)
	return exec.Command("bash", "-c", cmdStr)
}

func getExportKeyFlags(keyringDir string) string {
	return fmt.Sprintf(
		"--keyring-backend test --keyring-dir %s --unarmored-hex --unsafe",
		keyringDir,
	)
}
