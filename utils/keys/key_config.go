package keys

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
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
	kp := filepath.Join(home, kc.Dir)
	pterm.Info.Printfln("creating %s in %s", kc.ID, kp)
	args := []string{
		"keys", "add", kc.ID, "--keyring-backend", string(kc.KeyringBackend),
		"--keyring-dir", kp, "--home", kp,
		"--output", "json",
	}

	if kc.ShouldRecover {
		args = append(args, "--recover")
	}

	if kc.ShouldRecover {
		err := bash.ExecCommandWithInteractions(kc.ChainBinary, args...)
		if err != nil {
			return nil, err
		}

		ki, err := kc.Info(home)
		if err != nil {
			return nil, err
		}

		return ki, nil
	}

	out, err := RunCmdBasedOnKeyringBackend(home, kc.ChainBinary, args, kc.KeyringBackend)
	if err != nil {
		return nil, err
	}

	return ParseAddressFromOutput(out)
}

func (kc KeyConfig) Info(home string) (*KeyInfo, error) {
	kp := filepath.Join(home, kc.Dir)
	args := []string{
		"keys",
		"show",
		kc.ID,
		"--keyring-backend",
		string(kc.KeyringBackend),
		"--keyring-dir",
		kp,
		"--output",
		"json",
	}

	output, err := RunCmdBasedOnKeyringBackend(home, kc.ChainBinary, args, kc.KeyringBackend)
	if err != nil {
		return nil, err
	}

	return ParseAddressFromOutput(output)
}

func (kc KeyConfig) Address(home string) (string, error) {
	kp := filepath.Join(home, kc.Dir)
	args := []string{
		"keys",
		"show",
		kc.ID,
		"--address",
		"--keyring-backend",
		string(kc.KeyringBackend),
		"--keyring-dir",
		kp,
	}

	output, err := RunCmdBasedOnKeyringBackend(home, kc.ChainBinary, args, kc.KeyringBackend)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output.String()), nil
}

func (kc KeyConfig) IsInKeyring(
	home string,
) (bool, error) {
	kp := filepath.Join(home, kc.Dir)
	args := []string{
		"keys", "list", "--output", "json",
		"--keyring-backend", string(kc.KeyringBackend), "--keyring-dir", kp,
	}

	var ki []KeyInfo
	out, err := RunCmdBasedOnKeyringBackend(home, kc.ChainBinary, args, kc.KeyringBackend)
	if err != nil {
		return false, err
	}

	fmt.Println(out.String())
	if strings.Contains(out.String(), "No records were found in keyring") {
		return false, nil
	}

	err = json.Unmarshal(out.Bytes(), &ki)
	if err != nil {
		return false, err
	}

	if len(ki) == 0 {
		return false, nil
	}

	return slices.ContainsFunc(
		ki, func(i KeyInfo) bool {
			return strings.EqualFold(i.Name, kc.ID)
		},
	), nil
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
