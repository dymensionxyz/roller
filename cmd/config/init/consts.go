package initconfig

import "github.com/dymensionxyz/roller/cmd/utils"

var FlagNames = struct {
	Home          string
	TokenSupply   string
	RollappBinary string
	HubID         string
}{
	Home:          "home",
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
}

const StagingHubID = "internal-devnet"
const LocalHubID = "local"

var Hubs = map[string]utils.HubData{
	StagingHubID: {
		API_URL: "https://rest-hub-devnet.dymension.xyz",
		ID:      "devnet_666-1",
		RPC_URL: "https://rpc-hub-devnet.dymension.xyz:443",
	},
	LocalHubID: {
		API_URL: "http://localhost:1318",
		ID:      "dymension_100-1",
		RPC_URL: "http://localhost:36657",
	},
}
