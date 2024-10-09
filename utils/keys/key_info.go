package keys

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/utils/bash"
)

// KeyInfo struct stores information about a generated wallet
type KeyInfo struct {
	Name          string `json:"name"`
	Address       string `json:"address"`
	Mnemonic      string `json:"mnemonic"`
	PubKey        string `json:"pubkey"`
	PrintName     bool
	PrintMnemonic bool
	PrintPubKey   bool
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

func WithPubKey() KeyInfoOption {
	return func(opts *KeyInfo) {
		opts.PrintPubKey = true
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

	if ki.PrintPubKey {
		fmt.Printf("\t%s\n", ki.PubKey)
	}
	if ki.PrintMnemonic {
		fmt.Printf("\t%s\n", ki.Mnemonic)
		fmt.Println()
		fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))
	}

	fmt.Println()
}

func GetAddressInfoBinary(keyConfig KeyConfig, home string) (*KeyInfo, error) {
	showKeyCommand := exec.Command(
		keyConfig.ChainBinary,
		"keys",
		"show",
		keyConfig.ID,
		"--keyring-backend",
		"test",
		"--keyring-dir",
		filepath.Join(home, keyConfig.Dir),
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

func PrintAddressesWithTitle(addresses []KeyInfo) {
	pterm.DefaultSection.WithIndentCharacter("🔑").Println("Addresses")
	for _, address := range addresses {
		address.Print(WithMnemonic(), WithName())
	}
}
