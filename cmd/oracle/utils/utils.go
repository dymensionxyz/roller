package oracleutils

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/keys"
)

func GetOracleKeyConfig(a consts.VMType, ot string) ([]keys.KeyConfig, error) {
	fmt.Println("oracle type for key config", ot)
	kc := keys.KeyConfig{
		Dir:            filepath.Join(consts.ConfigDirName.Oracle, ot),
		ID:             consts.KeysIds.Oracle,
		ChainBinary:    consts.Executables.RollappEVM,
		Type:           consts.EVM_ROLLAPP,
		KeyringBackend: consts.SupportedKeyringBackends.Test,
	}

	var res *keys.KeyConfig
	var err error

	if a == consts.WASM_ROLLAPP {
		res, err = keys.NewKeyConfig(
			kc.Dir,
			kc.ID,
			kc.ChainBinary,
			kc.Type,
			kc.KeyringBackend,
			keys.WithCustomAlgo("secp256k1"),
		)
	} else {
		res, err = keys.NewKeyConfig(
			kc.Dir,
			kc.ID,
			kc.ChainBinary,
			kc.Type,
			kc.KeyringBackend,
		)
	}

	var keys []keys.KeyConfig

	if err != nil {
		return nil, err
	}

	keys = append(keys, *res)

	return keys, nil
}

func ExtractNetworkID(rollappID string) (int, error) {
	parts := strings.Split(rollappID, "_")
	if len(parts) != 2 {
		return 0, fmt.Errorf("invalid rollapp ID format")
	}

	middlePart := parts[1]
	epochIndex := strings.LastIndex(middlePart, "-")
	if epochIndex == -1 {
		return 0, fmt.Errorf("invalid rollapp ID format: missing epoch separator")
	}

	middlePart = middlePart[:epochIndex]
	value, err := strconv.Atoi(middlePart)
	if err != nil {
		return 0, fmt.Errorf("invalid rollapp ID format: %v", err)
	}

	return value, nil
}
