package utils

import (
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/bech32"
)

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
