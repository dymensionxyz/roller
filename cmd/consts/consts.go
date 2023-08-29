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
	Dymension:  fmt.Sprintf("%s/dymd", binsDir),
	Celestia:   fmt.Sprintf("%s/celestia", internalBinsDir),
	CelKey:     fmt.Sprintf("%s/cel-key", internalBinsDir),
	Relayer:    fmt.Sprintf("%s/rly", internalBinsDir),
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
