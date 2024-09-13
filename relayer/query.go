package relayer

import (
	"math/big"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
)

var oneDayRelayPrice = big.NewInt(1)

func getRolRlyAccData(cfg config.RollappConfig) (*utils.AccountData, error) {
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

func getHubRlyAccData(cfg config.RollappConfig) (*utils.AccountData, error) {
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

func GetRelayerAccountsData(cfg config.RollappConfig) ([]utils.AccountData, error) {
	var data []utils.AccountData

	// rollappRlyAcc, err := getRolRlyAccData(cfg)
	// if err != nil {
	// 	return nil, err
	// }
	// data = append(data, *rollappRlyAcc)

	hubRlyAcc, err := getHubRlyAccData(cfg)
	if err != nil {
		return nil, err
	}

	data = append(data, *hubRlyAcc)
	return data, nil
}

func GetRelayerInsufficientBalances(
	config config.RollappConfig,
) ([]utils.NotFundedAddressData, error) {
	var insufficientBalances []utils.NotFundedAddressData

	accData, err := GetRelayerAccountsData(config)
	if err != nil {
		return nil, err
	}

	for _, acc := range accData {
		if acc.Balance.Amount.Cmp(oneDayRelayPrice) < 0 {
			insufficientBalances = append(
				insufficientBalances, utils.NotFundedAddressData{
					KeyName:         consts.KeysIds.RollappRelayer,
					Address:         acc.Address,
					CurrentBalance:  acc.Balance.Amount,
					RequiredBalance: oneDayRelayPrice,
					Denom:           config.Denom,
					Network:         config.RollappID,
				},
			)
		}
	}

	return insufficientBalances, nil
}
