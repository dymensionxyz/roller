package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

// TODO: KeyInfo and AddressData seem redundant, should be moved into
// location
type KeyInfo struct {
	Address string `json:"address"`
}

func ParseAddressFromOutput(output bytes.Buffer) (string, error) {
	key := &KeyInfo{}
	err := json.Unmarshal(output.Bytes(), key)
	if err != nil {
		return "", err
	}
	return key.Address, nil
}

type SensitiveKeyInfo struct {
	Name     string `json:"name"`
	Address  string `json:"address"`
	Mnemonic string `json:"mnemonic"`
}

// TODO: the entire sensitive output thing could be simpler
func ParseAddressFromOutputWithSensisiveOutput(output bytes.Buffer) (*SensitiveKeyInfo, error) {
	key := &SensitiveKeyInfo{}
	err := json.Unmarshal(output.Bytes(), key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// KeyConfig struct store information about a wallet
// Dir refers to the keyringDir where the key is created
type KeyConfig struct {
	Dir string
	// TODO: this is not descriptive, Name would be more expressive
	ID          string
	ChainBinary string
	Type        config.VMType
}

func GetAddressBinary(keyConfig KeyConfig, binaryPath string) (string, error) {
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
		return "", err
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

func PrintAddressesWithTitle(addresses []AddressData) {
	fmt.Printf("🔑 Addresses:\n\n")
	PrintAddresses(addresses)
}

func PrintSecretAddressesWithTitle(addresses []SecretAddressData) {
	fmt.Printf("🔑 Addresses:\n\n")
	PrintSecretAddresses(addresses)
}

func PrintAddresses(addresses []AddressData) {
	data := make([][]string, 0, len(addresses))
	for _, address := range addresses {
		data = append(data, []string{address.Name, address.Addr})
	}
	table := tablewriter.NewWriter(os.Stdout)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	table.SetBorder(false)
	table.AppendBulk(data)
	table.Render()
}

type SecretAddressData struct {
	AddressData
	Mnemonic string
}

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
