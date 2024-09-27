package utils

import (
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
)

type keyConfigOptions struct {
	recover bool
}

type KeyConfigOption func(opt *keyConfigOptions) error

// TODO: add the keyring to use for the key
// KeyConfig struct store information about a wallet
// Dir refers to the keyringDir where the key is created
type KeyConfig struct {
	Dir string
	// TODO: this is not descriptive, Name would be more expressive
	ID            string
	ChainBinary   string
	Type          consts.VMType
	ShouldRecover bool
}

func WithRecover() KeyConfigOption {
	return func(options *keyConfigOptions) error {
		options.recover = true
		return nil
	}
}

func NewKeyConfig(
	dir, id, cb string,
	vmt consts.VMType,
	opts ...KeyConfigOption,
) (*KeyConfig, error) {
	var options keyConfigOptions

	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	shouldRecover := options.recover

	return &KeyConfig{
		Dir:           dir,
		ID:            id,
		ChainBinary:   cb,
		Type:          vmt,
		ShouldRecover: shouldRecover,
	}, nil
}

func (kc KeyConfig) Create(home string) (*KeyInfo, error) {
	args := []string{
		"keys", "add", kc.ID, "--keyring-backend", "test",
		"--keyring-dir", filepath.Join(home, kc.Dir),
		"--output", "json",
	}

	if kc.ShouldRecover {
		args = append(args, "--recover")
	}
	createKeyCommand := exec.Command(kc.ChainBinary, args...)

	if kc.ShouldRecover {
		err := bash.ExecCommandWithInteractions(kc.ChainBinary, args...)
		if err != nil {
			return nil, err
		}

		return nil, errors.New("forced stop")
	}
	out, err := bash.ExecCommandWithStdout(createKeyCommand)
	if err != nil {
		return nil, err
	}
	return ParseAddressFromOutput(out)
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
	out, err := bash.ExecCommandWithStdout(showKeyCmd)
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
	out, err := bash.ExecCommandWithStdout(cmd)
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
