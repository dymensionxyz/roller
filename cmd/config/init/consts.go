package initconfig

import "github.com/dymensionxyz/roller/config"

var FlagNames = struct {
	TokenSupply   string
	RollappBinary string
	HubID         string
	Decimals      string
	Interactive   string
	DAType        string
	VMType        string
}{
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
	Interactive:   "interactive",
	Decimals:      "decimals",
	DAType:        "da",
	VMType:        "vm-type",
}

const (
	StagingHubName = "devnet"
	LocalHubName   = "local"
)

// TODO(#112): The avaialble hub networks should be read from YAML file
var Hubs = map[string]config.HubData{
	StagingHubName: {
		API_URL:   "https://dymension-devnet.api.silknodes.io:443",
		ID:        "devnet_304-1",
		RPC_URL:   "https://dymension-devnet.rpc.silknodes.io:443",
		GAS_PRICE: "0.25",
	},
	LocalHubName: {
		API_URL:   "http://localhost:1318",
		ID:        "dymension_100-1",
		RPC_URL:   "http://localhost:36657",
		GAS_PRICE: "0",
	},
}
