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
	NoOutput      string
}{
	TokenSupply:   "token-supply",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
	Interactive:   "interactive",
	Decimals:      "decimals",
	DAType:        "da",
	VMType:        "vm-type",
	NoOutput:      "no-output",
}

const (
	StagingHubName    = "devnet"
	FroopylandHubName = "froopyland"
	LocalHubName      = "local"
	LocalHubID        = "dymension_100-1"
)

// TODO(#112): The avaialble hub networks should be read from YAML file
var Hubs = map[string]config.HubData{
	StagingHubName: {
		API_URL:         "https://dymension-devnet.api.silknodes.io:443",
		ID:              "devnet_304-1",
		RPC_URL:         "https://dymension-devnet.rpc.silknodes.io:443",
		ARCHIVE_RPC_URL: "https://dymension-devnet.rpc.silknodes.io:443",
		GAS_PRICE:       "0.25",
	},
	FroopylandHubName: {
		API_URL:         "https://froopyland.blockpi.network:443/lcd/v1/public",
		ID:              "froopyland_100-1",
		RPC_URL:         "https://froopyland.blockpi.network:443/rpc/v1/public",
		ARCHIVE_RPC_URL: "https://froopyland.blockpi.network:443/rpc/v1/public",
		GAS_PRICE:       "0.25",
	},
	LocalHubName: {
		API_URL:         "http://localhost:1318",
		ID:              LocalHubID,
		RPC_URL:         "http://localhost:36657",
		ARCHIVE_RPC_URL: "http://localhost:36657",
		GAS_PRICE:       "100000000",
	},
}
