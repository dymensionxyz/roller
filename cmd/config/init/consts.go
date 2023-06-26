package initconfig

import "github.com/dymensionxyz/roller/cmd/utils"

var FlagNames = struct {
	Home          string
	Decimals      string
	RollappBinary string
	HubID         string
}{
	Home:          "home",
	Decimals:      "decimals",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
}

const TestnetHubID = "35-C"
const StagingHubID = "internal-devnet"
const LocalHubID = "local"

var Hubs = map[string]utils.HubData{
	TestnetHubID: {
		API_URL: "https://rest-hub-35c.dymension.xyz",
		ID:      "35-C",
		RPC_URL: "https://rpc-hub-35c.dymension.xyz:443",
	},
	StagingHubID: {
		API_URL: "https://rest-hub-devnet.dymension.xyz",
		ID:      "internal-devnet",
		RPC_URL: "https://rpc-hub-devnet.dymension.xyz:443",
	},
	LocalHubID: {
		API_URL: "http://localhost:1317",
		ID:      "local-testnet",
		RPC_URL: "http://localhost:36657",
	},
}
