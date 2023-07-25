package utils

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os/exec"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
)

type ChainQueryConfig struct {
	Denom  string
	RPC    string
	Binary string
}

func QueryBalance(chainConfig ChainQueryConfig, address string) (Balance, error) {
	cmd := exec.Command(chainConfig.Binary, "query", "bank", "balances", address, "--node", chainConfig.RPC, "--output", "json")
	out, err := ExecBashCommand(cmd)
	if err != nil {
		return Balance{}, err
	}
	return ParseBalanceFromResponse(out, chainConfig.Denom)
}

func ParseBalance(balResp BalanceResp) (*big.Int, error) {
	amount := new(big.Int)
	_, ok := amount.SetString(balResp.Amount, 10)
	if !ok {
		return nil, errors.New("unable to convert balance amount to big.Int")
	}
	return amount, nil
}

func ParseBalanceFromResponse(out bytes.Buffer, denom string) (Balance, error) {
	var balanceResp BalanceResponse
	err := json.Unmarshal(out.Bytes(), &balanceResp)
	if err != nil {
		return Balance{}, err
	}

	balance := Balance{
		Denom:  denom,
		Amount: big.NewInt(0),
	}
	for _, resbalance := range balanceResp.Balances {
		if resbalance.Denom != denom {
			continue
		}
		amount, err := ParseBalance(resbalance)
		if err != nil {
			return Balance{}, err
		}
		balance.Amount = amount
	}
	return balance, nil
}

type Balance struct {
	Denom  string   `json:"denom"`
	Amount *big.Int `json:"amount"`
}

func (b *Balance) String() string {
	return b.Amount.String() + b.Denom
}

func formatBalance(balance *big.Int, decimals uint) string {
	divisor := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(decimals)), nil)
	quotient := new(big.Int).Div(balance, divisor)
	remainder := new(big.Int).Mod(balance, divisor)
	remainderStr := fmt.Sprintf("%0*s", decimals, remainder.String())
	const decimalPrecision = 6
	if len(remainderStr) > decimalPrecision {
		remainderStr = remainderStr[:decimalPrecision]
	}
	remainderStr = strings.TrimRight(remainderStr, "0")
	if remainderStr != "" {
		return fmt.Sprintf("%s.%s", quotient.String(), remainderStr)
	}
	return quotient.String()
}

func (b *Balance) BiggerDenomStr(cfg config.RollappConfig) string {
	biggerDenom := b.Denom[1:]
	decimalsMap := map[string]uint{
		consts.Denoms.Hub:      18,
		consts.Denoms.Celestia: 6,
		consts.Denoms.Avail:    18,
		cfg.Denom:              cfg.Decimals,
	}
	decimals, _ := decimalsMap[b.Denom]
	formattedBalance := formatBalance(b.Amount, decimals)
	return formattedBalance + strings.ToUpper(biggerDenom)
}

type BalanceResp struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
type BalanceResponse struct {
	Balances []BalanceResp `json:"balances"`
}

type AccountData struct {
	Address string
	Balance Balance
}

func GetSequencerInsufficientAddrs(cfg config.RollappConfig, requiredBalance *big.Int) ([]NotFundedAddressData, error) {
	sequencerData, err := GetSequencerData(cfg)
	if err != nil {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	for _, seq := range sequencerData {
		if seq.Balance.Amount.Cmp(requiredBalance) < 0 {
			return []NotFundedAddressData{
				{
					Address:         seq.Address,
					Denom:           consts.Denoms.Hub,
					CurrentBalance:  seq.Balance.Amount,
					RequiredBalance: requiredBalance,
					KeyName:         consts.KeysIds.HubSequencer,
					Network:         cfg.HubData.ID,
				},
			}, nil
		}
	}
	return []NotFundedAddressData{}, nil
}
