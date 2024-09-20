package query

import (
	"context"

	cosmosclient "github.com/cosmos/cosmos-sdk/client"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	eibctypes "github.com/dymensionxyz/dymension/v3/x/eibc/types"
)

// HubQueryClients is gRPC query client for Dymension Hub data.
type HubQueryClients struct {
	*GenericQueryClients
	EIbcQueryClient eibctypes.QueryClient
}

// RollAppQueryClients is gRPC query client for RollApp data.
// This project does not import rollapp so this struct is not used.
type RollAppQueryClients struct {
	*GenericQueryClients
}

// GenericQueryClients is gRPC query client for modules those exist in both Dymension Hub and RollApp.
type GenericQueryClients struct {
	AuthQueryClient authtypes.QueryClient
	BankQueryClient banktypes.QueryClient
}

// NewHubQueryClients initialize query clients for querying Dymension Hub data.
func NewHubQueryClients(hcc cosmosclient.Context) *HubQueryClients {
	return &HubQueryClients{
		GenericQueryClients: newGenericQueryClients(hcc),
		EIbcQueryClient:     eibctypes.NewQueryClient(hcc),
	}
}

func newGenericQueryClients(clientCtx cosmosclient.Context) *GenericQueryClients {
	return &GenericQueryClients{
		AuthQueryClient: authtypes.NewQueryClient(clientCtx),
		BankQueryClient: banktypes.NewQueryClient(clientCtx),
	}
}

func (q *HubQueryClients) QueryAllBalances(account string) (sdk.Coins, error) {
	res, err := q.BankQueryClient.AllBalances(context.Background(), &banktypes.QueryAllBalancesRequest{
		Address: account,
	})
	if err != nil {
		return nil, err
	}
	return res.Balances, nil
}
