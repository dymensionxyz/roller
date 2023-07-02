package initconfig

import "github.com/dymensionxyz/roller/cmd/utils"

var FlagNames = struct {
	TokenSupply   string
	RollappBinary string
	HubID         string
	Interactive   string
}{
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
	Interactive:   "interactive",
}

const (
	StagingHubID = "devnet"
	LocalHubID   = "local"
)

// TODO(#112): The avaialble hub networks should be read from YAML file
var Hubs = map[string]utils.HubData{
	StagingHubID: {
		API_URL: "https://dymension.devnet.api.silknodes.io:443",
		ID:      "devnet_304-1",
		RPC_URL: "https://dymension.devnet.rpc.silknodes.io:443",
	},
	LocalHubID: {
		API_URL: "http://localhost:1318",
		ID:      "dymension_100-1",
		RPC_URL: "http://localhost:36657",
	},
}
