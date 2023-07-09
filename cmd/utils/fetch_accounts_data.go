package utils

import (
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

func GetRelayerAccountsData(cfg config.RollappConfig) ([]AccountData, error) {
	data := []AccountData{{}, {}} // Initialize with two zero-value AccountData instances
	rollappRlyAcc, err := GetRolRlyAccData(cfg)
	if err != nil {
		return data, err
	}
	data[0] = *rollappRlyAcc // Update the first element
	hubRlyAcc, err := GetHubRlyAccData(cfg)
	if err != nil {
		return data, err
	}
	data[1] = *hubRlyAcc // Update the second element
	return data, nil
}
func GetRolRlyAccData(cfg config.RollappConfig) (*AccountData, error) {
	RollappRlyAddr, err := GetRelayerAddress(cfg.Home, cfg.RollappID)
	if err != nil {
		return nil, err
	}
	RollappRlyBalance, err := QueryBalance(ChainQueryConfig{
		RPC:    consts.DefaultRollappRPC,
		Denom:  cfg.Denom,
		Binary: cfg.RollappBinary,
	}, RollappRlyAddr)
	if err != nil {
		return nil, err
	}
	return &AccountData{
		Address: RollappRlyAddr,
		Balance: RollappRlyBalance,
	}, nil
}

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
	accData := &AccountData{} // Note: Using a pointer here
	sequencerAddress, err := GetAddressBinary(KeyConfig{
		ID:  consts.KeysIds.HubSequencer,
		Dir: filepath.Join(cfg.Home, consts.ConfigDirName.HubKeys),
	}, consts.Executables.Dymension)
	if err != nil {
		return []AccountData{*accData}, err
	}
	accData.Address = sequencerAddress

	sequencerBalance, err := QueryBalance(ChainQueryConfig{
		Binary: consts.Executables.Dymension,
		Denom:  consts.Denoms.Hub,
		RPC:    cfg.HubData.RPC_URL,
	}, sequencerAddress)
	if err != nil {
		return []AccountData{*accData}, err
	}
	accData.Balance = sequencerBalance

	return []AccountData{*accData}, nil
}
