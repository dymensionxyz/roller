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
}{
	HubSequencer:     "hub_sequencer",
	RollappSequencer: "rollapp_sequencer",
	RollappRelayer:   "relayer-rollapp-key",
}

var configDirName = struct {
	Rollapp   string
	Relayer   string
	LightNode string
}{
	Rollapp:   ".rollapp",
	Relayer:   ".relayer",
	LightNode: ".light-node",
}

const hubRPC = "https://rpc-hub-35c.dymension.xyz:443"
const lightNodeEndpointFlag = "light-node-endpoint"

const evmCoinType uint32 = 60
const hubChainId = "35-C"
const relayerKeysDirName = "keys"
const cosmosDefaultCointype uint32 = 118
const celestia_executable_path = "/Users/itaylevy/go/bin/celestia"
