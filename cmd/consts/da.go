package consts

var DaAuthTokenType = struct {
	Admin string
	Read  string
}{
	Admin: "admin",
	Read:  "read",
}

const (
	DefaultCelestiaMochaRest    = "https://api.celestia-mocha.com"
	DefaultCelestiaMochaRPC     = "https://celestia-testnet-rpc.itrocket.net:443"
	DefaultCelestiaMochaNetwork = "mocha-4"
	// https://docs.celestia.org/nodes/mocha-testnet#community-data-availability-da-grpc-endpoints-for-state-access
	DefaultCelestiaRest    = "https://api.celestia.pops.one"
	DefaultCelestiaRPC     = "http://rpc.celestia.pops.one:26657"
	DefaultCelestiaNetwork = "celestia"
)

type DAType string

const (
	Local    DAType = "mock"
	Celestia DAType = "celestia"
	Avail    DAType = "avail"
	LoadNetwork DAType = "loadnetwork"
	Sui      DAType = "sui"
	Mock     DAType = "mock"
)

type DaNetwork string

const (
	MockDA          DaNetwork = "mock"
	CelestiaTestnet DaNetwork = "mocha-4"
	CelestiaMainnet DaNetwork = "celestia"
	AvailTestnet    DaNetwork = "avail"
	AvailMainnet    DaNetwork = "avail-1" // change this with correct mainnet id
	LoadNetworkTestnet DaNetwork = "alphanet"
	LoadNetworkMainnet DaNetwork = "loadnetwork" // change this with correct mainnet id
	SuiTestnet      DaNetwork = "testnet"
	SuiMainnet      DaNetwork = "mainnet"
)

var DaNetworks = map[string]DaData{
	string(MockDA): {
		Backend:          Local,
		ApiUrl:           "",
		ID:               "mock",
		RpcUrl:           "",
		CurrentStateNode: "mock",
		StateNodes: []string{
			"mock1",
			"mock",
		},
		GasPrice: "",
	},
	string(CelestiaTestnet): {
		Backend:          Celestia,
		ApiUrl:           DefaultCelestiaMochaRest,
		ID:               CelestiaTestnet,
		RpcUrl:           DefaultCelestiaMochaRPC,
		CurrentStateNode: "rpc-mocha.pops.one",
		StateNodes: []string{
			"public-celestia-mocha4-consensus.numia.xyz",
			"full.consensus.mocha-4.celestia-mocha.com",
			"consensus-full-mocha-4.celestia-mocha.com",
			"rpc-mocha.pops.one",
		},
		GasPrice: "0.02",
	},
	string(CelestiaMainnet): {
		Backend:          Celestia,
		ApiUrl:           DefaultCelestiaRest,
		ID:               CelestiaMainnet,
		RpcUrl:           DefaultCelestiaRPC,
		CurrentStateNode: "rpc.celestia.pops.one",
		StateNodes: []string{
			"rpc-celestia.alphab.ai",
			"celestia.rpc.kjnodes.com",
		},
		GasPrice: "0.002",
	},
	string(AvailTestnet): {
		Backend:          Avail,
		ApiUrl:           "https://turing-rpc.avail.so/rpc",
		ID:               AvailTestnet,
		RpcUrl:           "wss://turing-rpc.avail.so/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(AvailMainnet): {
		Backend:          Avail,
		ApiUrl:           "https://mainnet-rpc.avail.so/rpc:443",
		ID:               AvailMainnet,
		RpcUrl:           "wss://mainnet.avail-rpc.com/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(LoadNetworkTestnet): {
		Backend:          LoadNetwork,
		ApiUrl:           "https://alphanet.load.network",
		ID:               LoadNetworkTestnet,
		RpcUrl:           "wss://alphanet.load.network/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(LoadNetworkMainnet): {
		Backend:          LoadNetwork,
		ApiUrl:           "",
		ID:               LoadNetworkMainnet,
		RpcUrl:           "",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(SuiTestnet): {
		Backend:          Sui,
		ApiUrl:           "",
		ID:               SuiTestnet,
		RpcUrl:           "https://fullnode.testnet.sui.io:443",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(SuiMainnet): {
		Backend:          Sui,
		ApiUrl:           "",
		ID:               SuiMainnet,
		RpcUrl:           "https://fullnode.mainnet.sui.io:443",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
}
