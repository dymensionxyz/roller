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

var MainnetHubData = config.HubData{
	API_URL:         "https://froopyland.blockpi.network:443/lcd/v1/public",
	ID:              MainnetHubID,
	RPC_URL:         "https://froopyland.blockpi.network:443/rpc/v1/public",
	ARCHIVE_RPC_URL: "https://froopyland.blockpi.network:443/rpc/v1/public",
	GAS_PRICE:       "20000000000",
}

var TestnetHubData = config.HubData{
	API_URL:         "https://froopyland.blockpi.network:443/lcd/v1/public",
	ID:              TestnetHubID,
	RPC_URL:         "https://froopyland.blockpi.network:443/rpc/v1/public",
	ARCHIVE_RPC_URL: "https://froopyland.blockpi.network:443/rpc/v1/public",
	GAS_PRICE:       "20000000000",
}

var LocalHubData = config.HubData{
	API_URL:         "http://localhost:1318",
	ID:              LocalHubID,
	RPC_URL:         "http://localhost:36657",
	ARCHIVE_RPC_URL: "http://localhost:36657",
	GAS_PRICE:       "100000000",
	SEQ_MIN_BOND:    "100dym",
}

// TODO(#112): The available hub networks should be read from YAML file
var Hubs = map[string]config.HubData{
	LocalHubName:   LocalHubData,
	TestnetHubName: TestnetHubData,
	MainnetHubName: MainnetHubData,
}

const (
	LocalHubName   = "local"
	TestnetHubName = "testnet"
	MainnetHubName = "mainnet"
)

const (
	LocalHubID   = "dymension_100-1"
	TestnetHubID = "blumbus_111-1"
	MainnetHubID = "dymension_1100-1"
)
