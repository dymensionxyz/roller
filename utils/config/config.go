package config

import (
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
)

type PathValue struct {
	Path  string
	Value interface{}
}

func GenerateCustomHubData() consts.HubData {
	id, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide hub chain id").Show()
	rpcUrl, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"provide hub rpc endpoint (including port, example: http://dym.dev:26657)",
	).Show()
	restUrl, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
		"provide hub rest api endpoint (including port, example: http://dym.dev:1318)",
	).Show()
	gasPrice, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide gas price").
		WithDefaultValue("2000000000").Show()

	id = strings.TrimSpace(id)
	rpcUrl = strings.TrimSpace(rpcUrl)
	restUrl = strings.TrimSpace(restUrl)
	gasPrice = strings.TrimSpace(gasPrice)

	hd := consts.HubData{
		API_URL:         restUrl,
		ID:              id,
		RPC_URL:         rpcUrl,
		ARCHIVE_RPC_URL: rpcUrl,
		GAS_PRICE:       gasPrice,
	}

	return hd
}
