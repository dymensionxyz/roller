package consts

import "fmt"

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
	Hub     string
	Rollapp string
	DA      string
}{
	Rollapp: "ethm",
	Hub:     "dym",
	DA:      "celestia",
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

var CoinTypes = struct {
	Cosmos uint32
	EVM    uint32
}{
	Cosmos: 118,
	EVM:    60,
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

const KeysDirName = "keys"
const DefaultRelayerPath = "hub-rollapp"
const DefaultRollappRPC = "http://localhost:26657"
const DefaultDALCRPC = "http://localhost:26659"
const CelestiaRestApiEndpoint = "https://api-arabica-8.consensus.celestia-arabica.com"
const DefaultCelestiaRPC = "consensus-full-arabica-8.celestia-arabica.com"
