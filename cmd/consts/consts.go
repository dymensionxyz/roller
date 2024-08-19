package consts

import (
	"fmt"
)

const (
	binsDir            = "/usr/local/bin"
	DefaultTokenSupply = "1000000000000000000000000000"
	DefaultFee         = 200000000000000000 // 0.2
)

var internalBinsDir = fmt.Sprintf("%s/roller_bins", binsDir)

var Executables = struct {
	Celestia    string
	RollappEVM  string
	Relayer     string
	Dymension   string
	CelKey      string
	Roller      string
	Simd        string
	Eibc        string
	CelestiaApp string
}{
	Roller:      fmt.Sprintf("%s/roller", binsDir),
	RollappEVM:  fmt.Sprintf("%s/rollapp-evm", binsDir),
	Dymension:   fmt.Sprintf("%s/dymd", binsDir),
	Celestia:    fmt.Sprintf("%s/celestia", internalBinsDir),
	CelKey:      fmt.Sprintf("%s/cel-key", internalBinsDir),
	Relayer:     fmt.Sprintf("%s/rly", internalBinsDir),
	Simd:        fmt.Sprintf("%s/simd", internalBinsDir),
	Eibc:        fmt.Sprintf("%s/eibc-client", binsDir),
	CelestiaApp: fmt.Sprintf("%s/celestia-appd", internalBinsDir),
}

var KeysIds = struct {
	HubSequencer     string
	HubGenesis       string
	RollappSequencer string
	RollappRelayer   string
	HubRelayer       string
	Celestia         string
	Eibc             string
}{
	HubSequencer:     "hub_sequencer",
	HubGenesis:       "hub_genesis",
	RollappSequencer: "rollapp_genesis_account",
	RollappRelayer:   "relayer-rollapp-key",
	HubRelayer:       "relayer-hub-key",
	Celestia:         "my_celes_key",
	Eibc:             "whale",
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
	Eibc        string
}{
	Rollapp:     "rollapp",
	Relayer:     "relayer",
	DALightNode: "da-light-node",
	HubKeys:     "hub-keys",
	LocalHub:    "local-hub",
	Eibc:        ".eibc-client",
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
	DefaultRelayerPath = "hub-rollapp"
	DefaultRollappRPC  = "http://localhost:26657"
)

var SpinnerMsgs = struct {
	UniqueIdVerification string
	BalancesVerification string
}{
	UniqueIdVerification: " Verifying unique RollApp ID...\n",
	BalancesVerification: " Verifying balances...\n",
}

var NodeType = struct {
	Sequencer string
	FullNode  string
}{
	Sequencer: "sequencer",
	FullNode:  "fullnode",
}

const RollerConfigFileName = "roller.toml"

type VMType string

const (
	SDK_ROLLAPP VMType = "sdk"
	EVM_ROLLAPP VMType = "evm"
)
