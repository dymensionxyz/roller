package management_web_service

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"time"

	"github.com/cosmos/cosmos-sdk/types/query"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/pterm/pterm"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
)

func HandlePartialBalances(c *gin.Context) {
	w := wrapGin(c)
	cfg := w.Config()

	denomsMetadata := cache.GetDenomsMetadata()
	if len(denomsMetadata) == 0 {
		resDenomsMetadata, err := cfg.HubQueryClients.BankQueryClient.DenomsMetadata(context.Background(), &banktypes.QueryDenomsMetadataRequest{
			Pagination: &query.PageRequest{
				Limit: 500,
			},
		})
		if err != nil {
			pterm.Error.Println("failed to fetch denom metadata", err)
		} else if resDenomsMetadata != nil {
			for _, metadata := range resDenomsMetadata.Metadatas {
				denomsMetadata[metadata.Base] = metadata
			}
			cache.SetDenomsMetadata(denomsMetadata)
		}
	}

	var whaleBalances sdk.Coins
	var whaleBalancesStr []string
	var err error
	{
		// example load account balances
		whaleBalances, err = cfg.HubQueryClients.QueryAllBalances(cfg.WhaleAccount)
		if err != nil {
			panic(err)
		}
		whaleBalancesStr = make([]string, 0, len(whaleBalances))

		for _, balance := range whaleBalances {
			if metadata, ok := denomsMetadata[balance.Denom]; ok {
				if balance.IsPositive() {
					var highestExponent uint32
					for _, denomUnit := range metadata.DenomUnits {
						if denomUnit != nil && denomUnit.Exponent > highestExponent {
							highestExponent = denomUnit.Exponent
						}
					}

					if highestExponent > 0 {
						high, low := new(big.Int).QuoRem(balance.Amount.BigInt(), new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(highestExponent)), nil), new(big.Int))
						whaleBalancesStr = append(whaleBalancesStr, fmt.Sprintf("%s: %s.%s (raw: %s %s)", metadata.Display, high.String(), low.String(), balance.Amount.String(), balance.Denom))
						continue
					}
				}
			}

			// fallback to print raw
			whaleBalancesStr = append(whaleBalancesStr, fmt.Sprintf("(raw) %s: %s", balance.Denom, balance.Amount.String()))
		}
	}

	c.HTML(http.StatusOK, "partial_balances.tmpl", gin.H{
		"account":        cfg.WhaleAccount,
		"time":           time.Now().UTC().Format(time.DateTime),
		"whale_balances": whaleBalancesStr,
	})
}

func HandlePartialEIbcClientLog(c *gin.Context) {
	// w := wrapGin(c)
	// cfg := w.Config()

	logs := cache.GetEIbcClientLog()
	if len(logs) == 0 {
		logs = []string{"No logs available"}
	}

	c.HTML(http.StatusOK, "partial_eibc_client_log.tmpl", gin.H{
		"logs": logs,
	})
}
