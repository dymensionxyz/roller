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

	j, _ := json.MarshalIndent(kc, "", " ")
	fmt.Println(string(j))

	if kc.ShouldRecover {
		args = append(args, "--recover")
	}
	createKeyCommand := exec.Command(kc.ChainBinary, args...)

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

	if kc.KeyringBackend == consts.SupportedKeyringBackends.OS {
		pterm.Info.Println("handling os keyring key creation")
		raFp := filepath.Join(home, string(consts.OsKeyringPwdFilePaths.RollApp))
		psw, err := config.ReadFromFile(raFp)
		if err != nil {
			return nil, err
		}

		pr := map[string]string{
			"Enter keyring passphrase":    psw,
			"Re-enter keyring passphrase": psw,
		}

		out, err := bash.ExecuteCommandWithPrompts(kc.ChainBinary, args, pr)
		if err != nil {
			pterm.Error.Printf("Command failed: %v\n", err)
			return nil, err
		}
		pterm.Info.Println("inputs handled")

		return ParseAddressFromOutput(*out)
	}

	out, err := bash.ExecCommandWithStdout(createKeyCommand)
	if err != nil {
		return nil, err
	}
	return ParseAddressFromOutput(out)
}

func (kc KeyConfig) Info(home string) (*KeyInfo, error) {
	showKeyCommand := exec.Command(
		kc.ChainBinary,
		"keys",
		"show",
		kc.ID,
		"--keyring-backend",
		string(kc.KeyringBackend),
		"--keyring-dir",
		filepath.Join(home, kc.Dir),
		"--output",
		"json",
	)

	output, err := bash.ExecCommandWithStdout(showKeyCommand)
	if err != nil {
		return nil, err
	}

	return ParseAddressFromOutput(output)
}

func (kc KeyConfig) Address(home string) (string, error) {
	showKeyCommand := exec.Command(
		kc.ChainBinary,
		"keys",
		"show",
		kc.ID,
		"--address",
		"--keyring-backend",
		string(kc.KeyringBackend),
		"--keyring-dir",
		filepath.Join(home, kc.Dir),
	)

	output, err := bash.ExecCommandWithStdout(showKeyCommand)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output.String()), nil
}

func (kc KeyConfig) IsInKeyring(
	home string,
) (bool, error) {
	keyringDir := filepath.Join(home, kc.Dir)

	cmd := exec.Command(
		kc.ChainBinary,
		"keys", "list", "--output", "json",
		"--keyring-backend", string(kc.KeyringBackend), "--keyring-dir", keyringDir,
	)

	var ki []KeyInfo
	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return false, err
	}

	fmt.Println(out.String())

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
