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

type KeyInfo struct {
	Address string `json:"address"`
}

func ParseAddressFromOutput(output bytes.Buffer) (string, error) {
	var key = &KeyInfo{}
	err := json.Unmarshal(output.Bytes(), key)
	if err != nil {
		return "", err
	}
	return key.Address, nil
}

type KeyConfig struct {
	Dir         string
	ID          string
	ChainBinary string
}

func GetAddressBinary(keyConfig KeyConfig, binaryPath string) (string, error) {
	showKeyCommand := exec.Command(
		binaryPath, "keys", "show", keyConfig.ID, "--keyring-backend", "test", "--keyring-dir", keyConfig.Dir,
		"--output", "json",
	)
	output, err := ExecBashCommand(showKeyCommand)
	if err != nil {
		return "", err
	}
	return ParseAddressFromOutput(output)
}

func GetRelayerAddress(home string, chainID string) (string, error) {
	showKeyCmd := exec.Command(
		consts.Executables.Relayer, "keys", "show", chainID, "--home", filepath.Join(home, consts.ConfigDirName.Relayer),
	)
	out, err := ExecBashCommand(showKeyCmd)
	return strings.TrimSuffix(out.String(), "\n"), err
}

type AddressData struct {
	Name string
	Addr string
}

func PrintAddresses(addresses []AddressData) {
	fmt.Printf("🔑 Addresses:\n\n")
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
	cmd := exec.Command(binaryPath, "debug", "addr", "ffffffffffffff", "--log_format", "json")
	out, err := ExecBashCommand(cmd)
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
