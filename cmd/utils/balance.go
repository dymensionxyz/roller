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

func VerifySequencerBalance(rollappConfig RollappConfig, requiredBalance *big.Int, insufficientBalanceErrHandler func(string) error) error {
	verifyBalanceConfig := VerifyBalanceConfig{
		RequiredBalance:               requiredBalance,
		InsufficientBalanceErrHandler: insufficientBalanceErrHandler,
		KeyConfig: GetKeyConfig{
			ID:  consts.KeyNames.HubSequencer,
			Dir: filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
		},
		ChainConfig: ChainQueryConfig{
			Binary: consts.Executables.Dymension,
			Denom:  consts.HubDenom,
			RPC:    rollappConfig.HubData.RPC_URL,
		},
	}
	return VerifyBalance(verifyBalanceConfig)
}

type VerifyBalanceConfig struct {
	RequiredBalance               *big.Int
	InsufficientBalanceErrHandler func(string) error
	KeyConfig                     GetKeyConfig
	ChainConfig                   ChainQueryConfig
}

func VerifyBalance(verifyBalanceConfig VerifyBalanceConfig) error {
	sequencerAddress, err := GetAddressBinary(
		verifyBalanceConfig.KeyConfig,
		consts.Executables.Dymension,
	)
	if err != nil {
		return err
	}
	sequencerBalance, err := QueryBalance(verifyBalanceConfig.ChainConfig, sequencerAddress)
	if err != nil {
		return err
	}
	if sequencerBalance.Cmp(verifyBalanceConfig.RequiredBalance) < 0 {
		return verifyBalanceConfig.InsufficientBalanceErrHandler(sequencerAddress)
	}
	return nil
}
