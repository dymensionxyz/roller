package utils

import (
	"bytes"
	"encoding/json"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
)

type KeyInfo struct {
	Address string `json:"address"`
}

func GetCelestiaAddress(keyringDir string) (string, error) {
	cmd := exec.Command(
		consts.Executables.CelKey,
		"show", consts.KeyNames.DALightNode, "--node.type", "light", "--keyring-dir", keyringDir, "--keyring-backend", "test", "--output", "json",
	)

	var out bytes.Buffer
	cmd.Stdout = &out

	err := cmd.Run()
	if err != nil {
		return "", err
	}

	var key = &KeyInfo{} 
	err = json.Unmarshal(out.Bytes(), key)
	if err != nil {
		return "", err
	}

	return key.Address, nil
}
