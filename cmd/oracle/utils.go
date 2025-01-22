package oracle

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/keys"
)

func getOracleKeyConfig(a consts.VMType) ([]keys.KeyConfig, error) {
	kc := keys.KeyConfig{
		Dir:            consts.ConfigDirName.Oracle,
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
