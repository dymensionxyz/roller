package utils

import (
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

func GetHubRlyAccData(cfg config.RollappConfig) (*AccountData, error) {
	HubRlyAddr, err := GetRelayerAddress(cfg.Home, cfg.HubData.ID)
	if err != nil {
		return nil, err
	}
	HubRlyBalance, err := QueryBalance(ChainQueryConfig{
		RPC:    cfg.HubData.RPC_URL,
		Denom:  consts.Denoms.Hub,
		Binary: consts.Executables.Dymension,
	}, HubRlyAddr)
	if err != nil {
		return nil, err
	}
	return &AccountData{
		Address: HubRlyAddr,
		Balance: HubRlyBalance,
	}, nil
}

func GetSequencerData(cfg config.RollappConfig) ([]AccountData, error) {
	sequencerAddress, err := GetAddressBinary(KeyConfig{
		ID:  consts.KeysIds.HubSequencer,
		Dir: filepath.Join(cfg.Home, consts.ConfigDirName.HubKeys),
	}, consts.Executables.Dymension)
	if err != nil {
		return nil, err
	}

	sequencerBalance, err := QueryBalance(ChainQueryConfig{
		Binary: consts.Executables.Dymension,
		Denom:  consts.Denoms.Hub,
		RPC:    cfg.HubData.RPC_URL,
	}, sequencerAddress)
	if err != nil {
		return nil, err
	}
	return []AccountData{
		{
			Address: sequencerAddress,
			Balance: sequencerBalance,
		},
	}, nil
}
