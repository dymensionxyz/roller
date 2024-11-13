package eibc

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/keys"
)

func GetKeyConfig() *keys.KeyConfig {
	kc := keys.KeyConfig{
		Dir:            consts.ConfigDirName.Eibc,
		ID:             consts.KeysIds.Eibc,
		ChainBinary:    consts.Executables.Dymension,
		Type:           "",
		KeyringBackend: consts.SupportedKeyringBackends.Test,
	}

	return &kc
}
