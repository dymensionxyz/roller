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

type GetKeyConfig struct {
	Dir string
	ID  string
}

type CreateKeyConfig struct {
	Dir      string
	ID       string
	CoinType uint32
	Algo     string
	Prefix   string
}

func GetAddressBinary(keyConfig GetKeyConfig, binaryPath string) (string, error) {
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

func MergeMaps(map1, map2 map[string]string) map[string]string {
	result := make(map[string]string)
	for key, value := range map1 {
		result[key] = value
	}
	for key, value := range map2 {
		result[key] = value
	}

	return result
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
