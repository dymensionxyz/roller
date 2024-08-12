package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	"github.com/pterm/pterm"

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

func GetAddressBinary(keyConfig KeyConfig, binaryPath string) (string, error) {
	showKeyCommand := exec.Command(
		binaryPath,
		"keys",
		"show",
		keyConfig.ID,
		"--address",
		"--keyring-backend",
		"test",
		"--keyring-dir",
		keyConfig.Dir,
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

// TODO: remove this function as it's redundant to PrintAddressesWithTitle
func PrintSecretAddressesWithTitle(addresses []SecretAddressData) {
	fmt.Printf("🔑 Addresses:\n")
	PrintSecretAddresses(addresses)
}

// TODO: remove this struct as it's redundant to KeyInfo
type SecretAddressData struct {
	AddressData
	Mnemonic string
}

// TODO: remove this function as it's redundant to *KeyInfo.Print
func PrintSecretAddresses(addresses []SecretAddressData) {
	for _, address := range addresses {
		fmt.Println(address.AddressData.Name)
		fmt.Println(address.AddressData.Addr)
		fmt.Println(address.Mnemonic)
	}
}
