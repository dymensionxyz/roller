package utils

import (
	"bytes"
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

// KeyInfo struct stores information about a generated wallet
type KeyInfo struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	Mnemonic      string `json:"mnemonic"`
	PrintName     bool
	PrintMnemonic bool
}

type KeyInfoOption func(*KeyInfo)

func WithName() KeyInfoOption {
	return func(opts *KeyInfo) {
		opts.PrintName = true
	}
}

func WithMnemonic() KeyInfoOption {
	return func(opts *KeyInfo) {
		opts.PrintMnemonic = true
	}
}

func (ki *KeyInfo) Print(o ...KeyInfoOption) {
	for _, opt := range o {
		opt(ki)
	}

	if ki.PrintName {
		pterm.DefaultBasicText.Println(pterm.LightGreen(ki.Name))
	}

	fmt.Printf("\t%s\n", ki.Address)

	if ki.PrintMnemonic {
		fmt.Printf("\t%s\n", ki.Mnemonic)
		fmt.Println()
		fmt.Println(pterm.LightYellow("💡 save the information and keep it safe"))
	}

	fmt.Println()
}

func ParseAddressFromOutput(output bytes.Buffer) (*KeyInfo, error) {
	key := &KeyInfo{}
	err := json.Unmarshal(output.Bytes(), key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func GetAddressInfoBinary(keyConfig KeyConfig, binaryPath string) (*KeyInfo, error) {
	showKeyCommand := exec.Command(
		binaryPath,
		"keys",
		"show",
		keyConfig.ID,
		"--keyring-backend",
		"test",
		"--keyring-dir",
		keyConfig.Dir,
		"--output",
		"json",
	)
	output, err := bash.ExecCommandWithStdout(showKeyCommand)
	if err != nil {
		return nil, err
	}
	return ParseAddressFromOutput(output)
}

func GetAddressBinary(keyConfig KeyConfig, home string) (string, error) {
	showKeyCommand := exec.Command(
		keyConfig.ChainBinary,
		"keys",
		"show",
		keyConfig.ID,
		"--address",
		"--keyring-backend",
		"test",
		"--keyring-dir",
		filepath.Join(home, keyConfig.Dir),
	)

	output, err := bash.ExecCommandWithStdout(showKeyCommand)
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(output.String()), nil
}

// TODO: refactor into options, with title
func PrintAddressesWithTitle(addresses []KeyInfo) {
	pterm.DefaultSection.WithIndentCharacter("🔑").Println("Addresses")
	for _, address := range addresses {
		address.Print(WithMnemonic(), WithName())
	}
}

// TODO: refactor customkeyringdir into options?
func IsAddressWithNameInKeyring(
	info KeyConfig,
	home string,
) (bool, error) {
	keyringDir := filepath.Join(home, info.Dir)

	cmd := exec.Command(
		info.ChainBinary,
		"keys", "list", "--output", "json",
		"--keyring-backend", "test", "--keyring-dir", keyringDir,
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
			return strings.EqualFold(i.Name, info.ID)
		},
	), nil
}

func IsRlyAddressWithNameInKeyring(
	info KeyConfig,
	chainId string,
) (bool, error) {
	cmd := exec.Command(
		consts.Executables.Relayer,
		"keys", "list", chainId, "--home", info.Dir,
	)
	fmt.Println(cmd.String())

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return false, err
	}
	fmt.Println(out.String())

	if strings.Contains(out.String(), fmt.Sprintf("no keys found for chain %s", chainId)) {
		return false, nil
	} else {
		return true, nil
	}
}

// TODO: remove this struct as it's redundant to KeyInfo
type SecretAddressData struct {
	AddressData
	Mnemonic string
}
