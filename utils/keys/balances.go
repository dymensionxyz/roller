package keys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/olekukonko/tablewriter"
	"github.com/pkg/errors"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/utils/bash"
)

func PrintInsufficientBalancesIfAny(
	addressesData []NotFundedAddressData,
) error {
	if len(addressesData) == 0 {
		return nil
	}

	printAddresses := func() {
		data := make([][]string, len(addressesData))
		for i, addressData := range addressesData {
			data[i] = []string{
				addressData.KeyName,
				addressData.Address,
				addressData.CurrentBalance.String() + addressData.Denom,
				addressData.RequiredBalance.String() + addressData.Denom,
				addressData.Network,
			}
		}

		table := tablewriter.NewWriter(os.Stdout)
		table.SetHeader([]string{"Name", "Address", "Current", "Required", "Network"})
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetBorder(false)
		table.AppendBulk(data)
		fmt.Println()
		table.Render()
		fmt.Println()
	}

	pterm.DefaultSection.WithIndentCharacter("🔔").
		Println("Please fund the addresses below.")
	printAddresses()

	// TODO: to util
	proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
		WithDefaultText(
			"press 'y' when the wallets are funded",
		).Show()
	if !proceed {
		pterm.Info.Println("exiting")
		return errors.New("cancelled by user")
	}

	return nil
}

type NotFundedAddressData struct {
	KeyName         string
	Address         string
	CurrentBalance  *big.Int
	RequiredBalance *big.Int
	Denom           string
	Network         string
}

type ChainQueryConfig struct {
	Denom  string
	RPC    string
	Binary string
}

func QueryBalance(chainConfig ChainQueryConfig, address string) (*cosmossdktypes.Coin, error) {
	cmd := exec.Command(
		chainConfig.Binary,
		"query",
		"bank",
		"balances",
		address,
		"--node",
		chainConfig.RPC,
		"--output",
		"json",
	)
	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}
	return ParseBalanceFromResponse(*out, chainConfig.Denom)
}

type Erc20ModuleTokenPair struct {
	Erc20Address string `json:"erc20_address"`
	Denom        string `json:"denom"`
	Enabled      bool   `json:"enabled"`
}

func QueryErc20Balance(
	rest string,
) (erc20ModuleTokenPairs []Erc20ModuleTokenPair, statusCode int, err error) {
	var resp *http.Response
	resp, err = http.Get(rest + "/evmos/erc20/v1/token_pairs")
	if err != nil {
		err = errors.Wrap(err, "failed to fetch ERC-20 module token pairs")
		return
	}

	statusCode = resp.StatusCode

	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(
			"failed to fetch ERC-20 module token pairs! Status code: %d",
			resp.StatusCode,
		)
		return
	}

	type responseStruct struct {
		TokenPairs []Erc20ModuleTokenPair `json:"token_pairs"`
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	bz, err := io.ReadAll(resp.Body)
	if err != nil {
		err = errors.Wrap(err, "failed to read response body of ERC-20 module token pairs")
		return
	}

	var response responseStruct
	err = json.Unmarshal(bz, &response)
	if err != nil {
		err = errors.Wrap(err, "failed to unmarshal response body of ERC-20 module token pairs")
		return
	}

	erc20ModuleTokenPairs = response.TokenPairs
	return
}

func ParseBalanceFromResponse(out bytes.Buffer, denom string) (*cosmossdktypes.Coin, error) {
	var balanceResp BalancesResp
	err := json.Unmarshal(out.Bytes(), &balanceResp)
	if err != nil {
		return nil, err
	}

	balance := cosmossdktypes.Coin{
		Denom:  denom,
		Amount: cosmossdkmath.NewInt(0),
	}

	if len(balanceResp.Balances) == 0 {
		return &balance, nil
	}

	for _, resbalance := range balanceResp.Balances {
		if resbalance.Denom != denom {
			continue
		}

		balance = resbalance
	}

	return &balance, nil
}

type AccountData struct {
	Address string
	Balance cosmossdktypes.Coin
}

type BalancesResp struct {
	Balances []cosmossdktypes.Coin `json:"balances"`
}

// func GetSequencerInsufficientAddrs(
// 	cfg config.RollappConfig,
// 	requiredBalance *big.Int,
// ) ([]NotFundedAddressData, error) {
// 	sequencerData, err := sequencer.GetSequencerData(cfg)
// 	if err != nil {
// 		return nil, err
// 	}
// 	if err != nil {
// 		return nil, err
// 	}
// 	for _, seq := range sequencerData {
// 		if seq.Balance.Amount.Cmp(requiredBalance) < 0 {
// 			return []NotFundedAddressData{
// 				{
// 					Address:         seq.Address,
// 					Denom:           consts.Denoms.Hub,
// 					CurrentBalance:  seq.Balance.Amount,
// 					RequiredBalance: requiredBalance,
// 					KeyName:         consts.KeysIds.HubSequencer,
// 					Network:         cfg.HubData.ID,
// 				},
// 			}, nil
// 		}
// 	}
// 	return []NotFundedAddressData{}, nil
// }
