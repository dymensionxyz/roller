package initconfig

var FlagNames = struct {
	Home          string
	DAEndpoint    string
	Decimals      string
	RollappBinary string
	HubID         string
}{
	Home:          "home",
	DAEndpoint:    "data-availability-endpoint",
	Decimals:      "decimals",
	RollappBinary: "rollapp-binary",
	HubID:         "hub",
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

var AddressPrefixes = struct {
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

const TestnetHubID = "35-C"
const StagingHubID = "internal-devnet"
const LocalHubID = "local"

var Hubs = map[string]HubData{
	TestnetHubID: {
		API_URL: "https://rest-hub-35c.dymension.xyz",
		ID:      "35-C",
		RPC_URL: "https://rpc-hub-35c.dymension.xyz:443",
	},
	StagingHubID: {
		API_URL: "https://rest-hub-devnet.dymension.xyz",
		ID:      "internal-devnet",
		RPC_URL: "https://rpc-hub-devnet.dymension.xyz:443",
	},
	LocalHubID: {
		API_URL: "http://localhost:1317",
		ID:      "local",
		RPC_URL: "http://localhost:36657",
	},
}

type HubData = struct {
	API_URL string
	ID      string
	RPC_URL string
}

const defaultRollappRPC = "http://localhost:26657"

var CoinTypes = struct {
	Cosmos uint32
	EVM    uint32
}{
	Cosmos: 118,
	EVM:    60,
}

const KeysDirName = "keys"
const RollerConfigFileName = "config.toml"
