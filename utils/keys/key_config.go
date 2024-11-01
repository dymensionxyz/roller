package keys

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

type keyConfigOptions struct {
	recover bool
}

type KeyConfigOption func(opt *keyConfigOptions) error

// TODO: add the keyring to use for the key
// KeyConfig struct store information about a wallet
// Dir refers to the keyringDir where the key is created
type KeyConfig struct {
	Dir            string
	ID             string
	ChainBinary    string
	Type           consts.VMType
	KeyringBackend consts.SupportedKeyringBackend
	ShouldRecover  bool
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
	kb consts.SupportedKeyringBackend,
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
		Dir:            dir,
		ID:             id,
		ChainBinary:    cb,
		Type:           vmt,
		KeyringBackend: kb,
		ShouldRecover:  shouldRecover,
	}, nil
}

func (kc KeyConfig) Create(home string) (*KeyInfo, error) {
	args := []string{
		"keys", "add", kc.ID, "--keyring-backend", string(kc.KeyringBackend),
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

		ki, err := GetAddressInfoBinary(kc, home)
		if err != nil {
			return nil, err
		}

		return ki, nil
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

// TODO: refactor, user should approve rather then pipe to 'yes'
func GetExportKeyCmdBinary(keyID, keyringDir, binary, keyringBackend string) *exec.Cmd {
	flags := getExportKeyFlags(keyringDir, keyringBackend)
	var commandStr string
	if binary == consts.Executables.CelKey {
		commandStr = fmt.Sprintf("%s export %s %s", binary, keyID, flags)
	} else {
		commandStr = fmt.Sprintf("%s keys export %s %s", binary, keyID, flags)
	}
	cmdStr := fmt.Sprintf("yes | %s", commandStr)
	return exec.Command("bash", "-c", cmdStr)
}

func GetExportPrivKeyCmd(kc KeyConfig) *exec.Cmd {
	c := exec.Command(
		kc.ChainBinary, "keys", "export", kc.ID, "--keyring-backend", string(kc.KeyringBackend),
	)

	return c
}

func getExportKeyFlags(keyringDir, keyringBackend string) string {
	return fmt.Sprintf(
		"--keyring-backend %s --keyring-dir %s --unarmored-hex --unsafe",
		keyringBackend,
		keyringDir,
	)
}
