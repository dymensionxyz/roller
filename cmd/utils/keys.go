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

func GetCelestiaAddress(rollerRoot string) (string, error) {
	daKeysDir := filepath.Join(rollerRoot, consts.ConfigDirName.DALightNode, consts.KeysDirName)
	cmd := exec.Command(
		consts.Executables.CelKey, "show", consts.KeysIds.DALightNode, "--node.type", "light", "--keyring-dir",
		daKeysDir, "--keyring-backend", "test", "--output", "json",
	)
	output, err := ExecBashCommand(cmd)
	if err != nil {
		return "", err
	}
	address, err := ParseAddressFromOutput(output)
	return address, err
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
