package initconfig

import "github.com/dymensionxyz/roller/cmd/utils"

var FlagNames = struct {
	TokenSupply   string
	RollappBinary string
	HubID         string
	Decimals      string
}{
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
	Decimals:      "decimals",
}

const (
	StagingHubID = "internal-devnet"
	LocalHubID   = "local"
)

// TODO(#112): The avaialble hub networks should be read from YAML file
var Hubs = map[string]utils.HubData{
	StagingHubID: {
		ApiUrl: "https://rest-hub-devnet.dymension.xyz",
		ID:     "devnet_666-1",
		RpcUrl: "https://rpc-hub-devnet.dymension.xyz:443",
	},
	LocalHubID: {
		ApiUrl: "http://localhost:1318",
		ID:     "dymension_100-1",
		RpcUrl: "http://localhost:36657",
	},
}
