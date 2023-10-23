package consts

import (
	"fmt"
	"github.com/dymensionxyz/roller/config"
	"math/big"
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
	Hub:      "udym",
	Celestia: "utia",
	Avail:    "aAVL",
}

const (
	KeysDirName        = "keys"
	DefaultRelayerPath = "rollapp-hub"
	DefaultRollappRPC  = "http://localhost:26657"
)

// TODO: Check DA LC write price on arabica and update this value.
var OneDAWritePrice = big.NewInt(1)

var SpinnerMsgs = struct {
	UniqueIdVerification string
	BalancesVerification string
}{
	UniqueIdVerification: " Verifying unique RollApp ID...\n",
	BalancesVerification: " Verifying balances...\n",
}

// TODO(#112): The avaialble hub networks should be read from YAML file
var Hubs = map[string]config.HubData{
	StagingHubName: {
		API_URL:   "https://dymension-devnet.api.silknodes.io:443",
		ID:        StagingHubID,
		RPC_URL:   "https://dymension-devnet.rpc.silknodes.io:443",
		GAS_PRICE: "0.25",
	},
	FroopylandHubName: {
		API_URL:   "https://froopyland.api.silknodes.io:443",
		ID:        FroopylandHubID,
		RPC_URL:   "https://froopyland.rpc.silknodes.io:443",
		GAS_PRICE: "0.25",
	},
	LocalHubName: {
		API_URL:   "http://localhost:1318",
		ID:        LocalHubID,
		RPC_URL:   "http://localhost:36657",
		GAS_PRICE: "0",
	},
}

const (
	StagingHubName    = "devnet"
	FroopylandHubName = "froopyland"
	LocalHubName      = "local"
	LocalHubID        = "dymension_100-1"
	StagingHubID      = "devnet_304-1"
	FroopylandHubID   = "froopyland_100-1"
)
