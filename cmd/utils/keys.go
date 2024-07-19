package utils

import (
	"bytes"
	"encoding/json"
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

func GetAddressBinary(keyConfig KeyConfig, binaryPath string) (*KeyInfo, error) {
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
	output, err := ExecBashCommandWithStdout(showKeyCommand)
	if err != nil {
		return nil, err
	}
	return ParseAddressFromOutput(output)
}

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
	return strings.TrimSuffix(out.String(), "\n"), err
}

type AddressData struct {
	Name string
	Addr string
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
