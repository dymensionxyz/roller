package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/dymensionxyz/roller/cmd/consts"
	"math/big"
	"os/exec"
	"path/filepath"
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

func GetSequencerInsufficientAddrs(config RollappConfig, requiredBalance big.Int) ([]NotFundedAddressData, error) {
	sequencerAddress, err := GetAddressBinary(GetKeyConfig{
		ID:  consts.KeyNames.HubSequencer,
		Dir: filepath.Join(config.Home, consts.ConfigDirName.HubKeys),
	}, consts.Executables.Dymension)
	if err != nil {
		return nil, err
	}
	sequencerBalance, err := QueryBalance(ChainQueryConfig{
		Binary: consts.Executables.Dymension,
		Denom:  consts.Denoms.Hub,
		RPC:    config.HubData.RPC_URL,
	}, sequencerAddress)
	if err != nil {
		return nil, err
	}
	if sequencerBalance.Cmp(&requiredBalance) < 0 {
		return []NotFundedAddressData{
			{
				Address:         sequencerAddress,
				Denom:           consts.Denoms.Hub,
				CurrentBalance:  sequencerBalance,
				RequiredBalance: &requiredBalance,
				KeyName:         consts.KeyNames.HubSequencer,
			},
		}, nil
	}
	return []NotFundedAddressData{}, nil
}
