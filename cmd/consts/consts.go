package consts

import (
	"fmt"

	"github.com/dymensionxyz/roller/config"
)

const binsDir = "/usr/local/bin"

var internalBinsDir = fmt.Sprintf("%s/roller_bins", binsDir)

var Executables = struct {
	Celestia   string
	RollappEVM string
	Relayer    string
	Dymension  string
	CelKey     string
	Roller     string
	Simd       string
}{
	Roller:     fmt.Sprintf("%s/roller", binsDir),
	RollappEVM: fmt.Sprintf("%s/rollapp_evm", binsDir),
	Dymension:  fmt.Sprintf("%s/dymd", binsDir),
	Celestia:   fmt.Sprintf("%s/celestia", internalBinsDir),
	CelKey:     fmt.Sprintf("%s/cel-key", internalBinsDir),
	Relayer:    fmt.Sprintf("%s/rly", internalBinsDir),
	Simd:       fmt.Sprintf("%s/simd", internalBinsDir),
}

var KeysIds = struct {
	HubSequencer     string
	RollappSequencer string
	RollappRelayer   string
	HubRelayer       string
}{
	HubSequencer:     "hub_sequencer",
	RollappSequencer: "rollapp_sequencer",
	RollappRelayer:   "relayer-rollapp-key",
	HubRelayer:       "relayer-hub-key",
}

var AddressPrefixes = struct {
	Hub string
}{
	Hub: "dym",
}

var ConfigDirName = struct {
	Rollapp     string
	Relayer     string
	DALightNode string
	HubKeys     string
	LocalHub    string
}{
	Rollapp:     "rollapp",
	Relayer:     "relayer",
	DALightNode: "da-light-node",
	HubKeys:     "hub-keys",
	LocalHub:    "local-hub",
}

var Denoms = struct {
	Hub      string
	Celestia string
	Avail    string
}{
	Hub:      "adym",
	Celestia: "utia",
	Avail:    "aAVL",
}

const (
	KeysDirName        = "keys"
	DefaultRelayerPath = "rollapp-hub"
	DefaultRollappRPC  = "http://localhost:26657"
)

var SpinnerMsgs = struct {
	UniqueIdVerification string
	BalancesVerification string
}{
	UniqueIdVerification: " Verifying unique RollApp ID...\n",
	BalancesVerification: " Verifying balances...\n",
}

var FroopylandHubData = config.HubData{
	API_URL:         "https://froopyland.blockpi.network:443/lcd/v1/public",
	ID:              FroopylandHubID,
	RPC_URL:         "https://froopyland.blockpi.network:443/rpc/v1/public",
	ARCHIVE_RPC_URL: "https://froopyland.blockpi.network:443/rpc/v1/public",
	GAS_PRICE:       "20000000000",
}

// TODO(#112): The available hub networks should be read from YAML file
var Hubs = map[string]config.HubData{
	StagingHubName: {
		API_URL:         "https://dymension-devnet.api.silknodes.io:443",
		ID:              StagingHubID,
		RPC_URL:         "https://dymension-devnet.rpc.silknodes.io:443",
		ARCHIVE_RPC_URL: "https://dymension-devnet.rpc.silknodes.io:443",
		GAS_PRICE:       "20000000000",
	},
	FroopylandHubName: FroopylandHubData,
	LocalHubName: {
		API_URL:         "http://localhost:1318",
		ID:              LocalHubID,
		RPC_URL:         "http://localhost:36657",
		ARCHIVE_RPC_URL: "http://localhost:36657",
		GAS_PRICE:       "100000000",
	},
	// TODO: Add mainnet hub data
	MainnetHubName: FroopylandHubData,
}

const (
	StagingHubName    = "devnet"
	FroopylandHubName = "froopyland"
	LocalHubName      = "local"
	MainnetHubName    = "mainnet"
	LocalHubID        = "dymension_100-1"
	StagingHubID      = "devnet_304-1"
	FroopylandHubID   = "froopyland_100-1"
)
