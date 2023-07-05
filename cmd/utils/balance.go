package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"math/big"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

type ChainQueryConfig struct {
	Denom  string
	RPC    string
	Binary string
}

func QueryBalance(chainConfig ChainQueryConfig, address string) (*big.Int, error) {
	cmd := exec.Command(chainConfig.Binary, "query", "bank", "balances", address, "--node", chainConfig.RPC, "--output", "json")
	out, err := ExecBashCommand(cmd)
	if err != nil {
		return nil, err
	}
	return ParseBalanceFromResponse(out, chainConfig.Denom)
}

func ParseBalanceFromResponse(out bytes.Buffer, denom string) (*big.Int, error) {
	var balanceResp BalanceResponse
	err := json.Unmarshal(out.Bytes(), &balanceResp)
	if err != nil {
		return nil, err
	}
	for _, balance := range balanceResp.Balances {
		if balance.Denom == denom {
			amount := new(big.Int)
			_, ok := amount.SetString(balance.Amount, 10)
			if !ok {
				return nil, errors.New("unable to convert balance amount to big.Int")
			}
			return amount, nil
		}
	}
	return big.NewInt(0), nil
}

type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type BalanceResponse struct {
	Balances []Balance `json:"balances"`
}

type AccountData struct {
	Address string
	Balance *big.Int
}

func GetSequencerInsufficientAddrs(cfg config.RollappConfig, requiredBalance big.Int) ([]NotFundedAddressData, error) {
	sequencerData, err := GetSequencerData(cfg)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	for _, seq := range sequencerData {
		if seq.Balance.Cmp(&requiredBalance) < 0 {
			return []NotFundedAddressData{
				{
					Address:         seq.Address,
					Denom:           consts.Denoms.Hub,
					CurrentBalance:  seq.Balance,
					RequiredBalance: &requiredBalance,
					KeyName:         consts.KeysIds.HubSequencer,
					Network:         cfg.HubData.ID,
				},
			}, nil
		}
	}
	return []NotFundedAddressData{}, nil
}
