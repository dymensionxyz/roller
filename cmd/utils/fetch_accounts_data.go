package utils

import (
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

func GetSequencerData(cfg config.RollappConfig) ([]AccountData, error) {
	seqAddr, err := GetAddressBinary(
		KeyConfig{
			ID:  consts.KeysIds.HubSequencer,
			Dir: filepath.Join(cfg.Home, consts.ConfigDirName.HubKeys),
		}, consts.Executables.Dymension,
	)
	if err != nil {
		return nil, err
	}

	sequencerBalance, err := QueryBalance(
		ChainQueryConfig{
			Binary: consts.Executables.Dymension,
			Denom:  consts.Denoms.Hub,
			RPC:    cfg.HubData.RPC_URL,
		}, seqAddr,
	)
	if err != nil {
		return nil, err
	}
	return []AccountData{
		{
			Address: seqAddr,
			Balance: sequencerBalance,
		},
	}, nil
}
