package utils

import (
	"fmt"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

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

func GetSequencerData(cfg config.RollappConfig) (*AccountData, error) {
	sequencerAddress, err := GetAddressBinary(GetKeyConfig{
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
	return &AccountData{
		Address: sequencerAddress,
		Balance: sequencerBalance,
	}, nil
}

func GetCelLCAccData(rollappConfig config.RollappConfig) (*AccountData, error) {
	celAddress, err := GetCelestiaAddress(rollappConfig.Home)
	if err != nil {
		return nil, err
	}
	var restQueryUrl = fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		consts.CelestiaRestApiEndpoint, celAddress,
	)
	balancesJson, err := RestQueryJson(restQueryUrl)
	if err != nil {
		return nil, err
	}
	balance, err := ParseBalanceFromResponse(*balancesJson, consts.Denoms.Celestia)
	if err != nil {
		return nil, err
	}
	return &AccountData{
		Address: celAddress,
		Balance: balance,
	}, nil
}
