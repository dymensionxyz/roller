package initconfig

import "github.com/dymensionxyz/roller/cmd/utils"

var FlagNames = struct {
	TokenSupply   string
	RollappBinary string
	HubID         string
	Decimals      string
	Interactive   string
}{
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
	Interactive:   "interactive",
	Decimals:      "decimals",
}

const (
	StagingHubName = "devnet"
	LocalHubName   = "local"
)

// TODO(#112): The avaialble hub networks should be read from YAML file
var Hubs = map[string]utils.HubData{
	StagingHubName: {
		API_URL:     "https://dymension.devnet.api.silknodes.io:443",
		ID:          "devnet_304-1",
		RPC_URL:     "https://dymension.devnet.rpc.silknodes.io:443",
		DisplayName: StagingHubName,
	},
	LocalHubName: {
		API_URL:     "http://localhost:1318",
		ID:          "dymension_100-1",
		RPC_URL:     "http://localhost:36657",
		DisplayName: LocalHubName,
	},
}
