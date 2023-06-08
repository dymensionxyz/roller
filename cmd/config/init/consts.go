package initconfig

var FlagNames = struct {
	Home          string
	DAEndpoint    string
	Decimals      string
	RollappBinary string
	HubRPC        string
}{
	Home:          "home",
	DAEndpoint:    "data-availability-endpoint",
	Decimals:      "decimals",
	RollappBinary: "rollapp-binary",
	HubRPC:        "hub-rpc",
}

var KeyNames = struct {
	HubSequencer     string
	RollappSequencer string
	RollappRelayer   string
	DALightNode      string
	HubRelayer       string
}{
	HubSequencer:     "hub_sequencer",
	RollappSequencer: "rollapp_sequencer",
	RollappRelayer:   "relayer-rollapp-key",
	DALightNode:      "my-celes-key",
	HubRelayer:       "relayer-hub-key",
}

var addressPrefixes = struct {
	Hub     string
	Rollapp string
	DA      string
}{
	Rollapp: "rol",
	Hub:     "dym",
	DA:      "celestia",
}

var ConfigDirName = struct {
	Rollapp     string
	Relayer     string
	DALightNode string
}{
	Rollapp:     "rollapp",
	Relayer:     "relayer",
	DALightNode: "light-node",
}

var HubData = struct {
	API_URL string
	ID      string
	RPC_URL string
}{
	API_URL: "http://127.0.0.1:1317",
	// API_URL: "https://api-hub-35c.dymension.xyz",
	ID:      "35-C",
	RPC_URL: "https://rpc-hub-35c.dymension.xyz:443",
}

const defaultRollappRPC = "http://localhost:26657"

const evmCoinType uint32 = 60
const KeysDirName = "keys"
const cosmosDefaultCointype uint32 = 118
const celestiaExecutablePath = "/usr/local/bin/roller_bins/celestia"
const defaultRollappBinaryPath = "/usr/local/bin/roller_bins/rollapp_evm"
const relayerExecutablePath = "/usr/local/bin/roller_bins/rly"
