package management_web_service

import (
	"fmt"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
	"net/http"
	"time"
)

func HandlePartialBalances(c *gin.Context) {
	w := wrapGin(c)
	cfg := w.Config()

	var whaleBalances sdk.Coins
	var whaleBalancesStr []string
	var err error
	{
		// example load account balances
		whaleBalances, err = cfg.HubQueryClients.QueryAllBalances(cfg.WhaleAccount)
		if err != nil {
			panic(err)
		}
		for _, balance := range whaleBalances {
			whaleBalancesStr = append(whaleBalancesStr, fmt.Sprintf("%s: %s", balance.Denom, balance.Amount.String()))
		}
	}

	c.HTML(http.StatusOK, "partial_balances.tmpl", gin.H{
		"account":        cfg.WhaleAccount,
		"time":           time.Now().UTC().Format(time.DateTime),
		"whale_balances": whaleBalances,
	})
}
