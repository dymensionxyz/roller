package sequencer

import (
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

func GetRolRlyAccData(cfg config.RollappConfig) (*utils.AccountData, error) {
	RollappRlyAddr, err := utils.GetRelayerAddress(cfg.Home, cfg.RollappID)
	if err != nil {
		return nil, err
	}
	rollappRPCEndpoint, err := GetRPCEndpoint(cfg)
	if err != nil {
		return nil, err
	}
	RollappRlyBalance, err := utils.QueryBalance(utils.ChainQueryConfig{
		RPC:    rollappRPCEndpoint,
		Denom:  cfg.Denom,
		Binary: cfg.RollappBinary,
	}, RollappRlyAddr)
	if err != nil {
		return nil, err
	}
	return &utils.AccountData{
		Address: RollappRlyAddr,
		Balance: RollappRlyBalance,
	}, nil
}

func GetRelayerAccountsData(cfg config.RollappConfig) ([]utils.AccountData, error) {
	data := []utils.AccountData{}
	rollappRlyAcc, err := GetRolRlyAccData(cfg)
	if err != nil {
		return nil, err
	}
	data = append(data, *rollappRlyAcc)
	hubRlyAcc, err := utils.GetHubRlyAccData(cfg)
	if err != nil {
		return nil, err
	}
	data = append(data, *hubRlyAcc)
	return data, nil
}
