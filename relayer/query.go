package relayer

import (
	cosmossdkmath "cosmossdk.io/math"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
)

var oneDayRelayPrice, _ = cosmossdkmath.NewIntFromString(
	"2000000000000000000",
) // 2000000000000000000 = 2dym

// TODO: refactor to use consts.RollappData
// nolint: unused
func getRolRlyAccData(home string, raData config.RollappConfig) (*utils.AccountData, error) {
	RollappRlyAddr, err := utils.GetRelayerAddress(home, raData.RollappID)
	seq := sequencer.GetInstance(raData)
	if err != nil {
		return nil, err
	}

	RollappRlyBalance, err := utils.QueryBalance(
		utils.ChainQueryConfig{
			RPC:    seq.GetRPCEndpoint(),
			Denom:  raData.Denom,
			Binary: consts.Executables.RollappEVM,
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

func getHubRlyAccData(home string, hd consts.HubData) (*utils.AccountData, error) {
	HubRlyAddr, err := utils.GetRelayerAddress(home, hd.ID)
	if err != nil {
		return nil, err
	}

	HubRlyBalance, err := utils.QueryBalance(
		utils.ChainQueryConfig{
			RPC:    hd.RPC_URL,
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

func GetRelayerAccountsData(
	home string,
	raData consts.RollappData,
	hd consts.HubData,
) ([]utils.AccountData, error) {
	var data []utils.AccountData

	// rollappRlyAcc, err := getRolRlyAccData(cfg)
	// if err != nil {
	// 	return nil, err
	// }
	// data = append(data, *rollappRlyAcc)

	hubRlyAcc, err := getHubRlyAccData(home, hd)
	if err != nil {
		return nil, err
	}

	data = append(data, *hubRlyAcc)
	return data, nil
}

func GetRelayerInsufficientBalances(
	raData consts.RollappData,
	hd consts.HubData,
) ([]utils.NotFundedAddressData, error) {
	var insufficientBalances []utils.NotFundedAddressData
	home := utils.GetRollerRootDir()

	accData, err := GetRelayerAccountsData(home, raData, hd)
	if err != nil {
		return nil, err
	}

	// consts.Denoms.Hub is used here because as of @202409 we no longer require rollapp
	// relayer account funding to establish IBC connection.
	for _, acc := range accData {
		if acc.Balance.Amount.Cmp(oneDayRelayPrice.BigInt()) < 0 {
			insufficientBalances = append(
				insufficientBalances, utils.NotFundedAddressData{
					KeyName:         consts.KeysIds.HubRelayer,
					Address:         acc.Address,
					CurrentBalance:  acc.Balance.Amount,
					RequiredBalance: oneDayRelayPrice.BigInt(),
					Denom:           consts.Denoms.Hub,
					Network:         hd.ID,
				},
			)
		}
	}

	return insufficientBalances, nil
}
