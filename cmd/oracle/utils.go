package oracle

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/keys"
)

func getOracleKeyConfig() ([]keys.KeyConfig, error) {
	kc := keys.KeyConfig{
		Dir:            consts.ConfigDirName.Oracle,
		ID:             consts.KeysIds.Oracle,
		ChainBinary:    consts.Executables.RollappEVM,
		Type:           consts.SDK_ROLLAPP,
		KeyringBackend: consts.SupportedKeyringBackends.Test,
	}

	res, err := keys.NewKeyConfig(
		kc.Dir,
		kc.ID,
		kc.ChainBinary,
		kc.Type,
		kc.KeyringBackend,
		keys.WithCustomAlgo("secp256k1"),
	)

	var keys []keys.KeyConfig

	if err != nil {
		return nil, err
	}

	keys = append(keys, *res)

	return keys, nil
}
