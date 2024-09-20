package management_web_service

import (
	"fmt"
	"net/http"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/gin-gonic/gin"
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
