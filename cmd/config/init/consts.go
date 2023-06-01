package init

var flagNames = struct {
	LightNodeEndpoint string
	Denom             string
	KeyPrefix         string
	Decimals          string
	RollappBinary     string
	HubRPC            string
}{
	LightNodeEndpoint: "light-node-endpoint",
	Denom:             "denom",
	KeyPrefix:         "key-prefix",
	Decimals:          "decimals",
	RollappBinary:     "rollapp-binary",
	HubRPC:            "hub-rpc",
}

var keyNames = struct {
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

var keyPrefixes = struct {
	Hub    string
	DA    string
}{
	Hub:    "dym",
	DA:    "celestia",
}

var configDirName = struct {
	Rollapp     string
	Relayer     string
	DALightNode string
}{
	Rollapp:     ".rollapp",
	Relayer:     "relayer",
	DALightNode: ".light-node",
}

const defaultHubRPC = "https://rpc-hub-35c.dymension.xyz:443"
const defaultRollappRPC = "http://localhost:26657"
const lightNodeEndpointFlag = "light-node-endpoint"

const evmCoinType uint32 = 60
const defaultHubId = "35-C"
const relayerKeysDirName = "keys"
const cosmosDefaultCointype uint32 = 118
const celestiaExecutablePath = "/usr/local/bin/roller/lib/celestia"
const defaultRollappBinaryPath = "/usr/local/bin/roller/lib/rollapp_evm"
const relayerExecutablePath = "/usr/local/bin/roller/lib/rly"
