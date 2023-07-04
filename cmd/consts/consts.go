package consts

import (
	"fmt"
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
}{
	Roller:     fmt.Sprintf("%s/roller", binsDir),
	RollappEVM: fmt.Sprintf("%s/rollapp_evm", binsDir),
	Celestia:   fmt.Sprintf("%s/celestia", internalBinsDir),
	CelKey:     fmt.Sprintf("%s/cel-key", internalBinsDir),
	Relayer:    fmt.Sprintf("%s/rly", internalBinsDir),
	Dymension:  fmt.Sprintf("%s/dymd", internalBinsDir),
}

var KeysIds = struct {
	HubSequencer     string
	RollappSequencer string
	RollappRelayer   string
	DALightNode      string
	HubRelayer       string
}{
	HubSequencer:     "hub_sequencer",
	RollappSequencer: "rollapp_sequencer",
	RollappRelayer:   "relayer-rollapp-key",
	DALightNode:      "my_celes_key",
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
}{
	Rollapp:     "rollapp",
	Relayer:     "relayer",
	DALightNode: "da-light-node",
	HubKeys:     "hub-keys",
}

var AlgoTypes = struct {
	Secp256k1    string
	Ethsecp256k1 string
}{
	Secp256k1:    "secp256k1",
	Ethsecp256k1: "eth_secp256k1",
}

var Denoms = struct {
	Hub      string
	Celestia string
}{
	Hub:      "udym",
	Celestia: "utia",
}

const (
	KeysDirName             = "keys"
	DefaultRelayerPath      = "hub-rollapp"
	DefaultRollappRPC       = "http://localhost:26657"
	DefaultDALCRPC          = "http://localhost:26659"
	CelestiaRestApiEndpoint = "https://api-mocha.pops.one"
	DefaultCelestiaRPC      = "rpc-mocha.pops.one"
	DefaultCeletiaNetowrk   = "mocha"
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
