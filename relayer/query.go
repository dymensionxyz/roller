package relayer

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
)

func GetRolRlyAccData(cfg config.RollappConfig) (*utils.AccountData, error) {
	RollappRlyAddr, err := utils.GetRelayerAddress(cfg.Home, cfg.RollappID)
	seq := sequencer.GetInstance(cfg)
	if err != nil {
		return nil, err
	}

	RollappRlyBalance, err := utils.QueryBalance(
		utils.ChainQueryConfig{
			RPC:    seq.GetRPCEndpoint(),
			Denom:  cfg.BaseDenom,
			Binary: cfg.RollappBinary,
		}, RollappRlyAddr,
	)
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
	hubRlyAcc, err := GetHubRlyAccData(cfg)
	if err != nil {
		return nil, err
	}

	data = append(data, *hubRlyAcc)
	return data, nil
}

func GetHubRlyAccData(cfg config.RollappConfig) (*utils.AccountData, error) {
	HubRlyAddr, err := utils.GetRelayerAddress(cfg.Home, cfg.HubData.ID)
	if err != nil {
		return nil, err
	}

	HubRlyBalance, err := utils.QueryBalance(
		utils.ChainQueryConfig{
			RPC:    cfg.HubData.RPC_URL,
			Denom:  consts.Denoms.Hub,
			Binary: consts.Executables.Dymension,
		}, HubRlyAddr,
	)
	if err != nil {
		return nil, err
	}

	return &utils.AccountData{
		Address: HubRlyAddr,
		Balance: HubRlyBalance,
	}, nil
}
