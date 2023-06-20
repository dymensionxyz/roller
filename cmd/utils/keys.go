package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"os/exec"

	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
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
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", errors.New(stderr.String())
	}

	var key = &KeyInfo{}
	err = json.Unmarshal(out.Bytes(), key)
	if err != nil {
		return "", err
	}

	return key.Address, nil
}

type KeyConfig struct {
	Dir      string
	ID       string
	CoinType uint32
	Prefix   string
}

func KeyInfoToBech32Address(info keyring.Info, prefix string) (string, error) {
	pk := info.GetPubKey()
	addr := types.AccAddress(pk.Address())
	bech32Address, err := bech32.ConvertAndEncode(prefix, addr.Bytes())
	if err != nil {
		return "", err
	}
	return bech32Address, nil
}

func GetAddress(keyConfig KeyConfig) (string, error) {
	kr, err := keyring.New(
		"",
		keyring.BackendTest,
		keyConfig.Dir,
		nil,
	)
	if err != nil {
		return "", err
	}
	keyInfo, err := kr.Key(keyConfig.ID)
	if err != nil {
		return "", err
	}
	formattedAddress, err := KeyInfoToBech32Address(keyInfo, keyConfig.Prefix)
	if err != nil {
		return "", err
	}
	return formattedAddress, nil
}
