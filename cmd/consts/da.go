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
)

type DaNetwork string

const (
	MockDA          DaNetwork = "mock"
	CelestiaTestnet DaNetwork = "mocha-4"
	CelestiaMainnet DaNetwork = "celestia"
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
}
