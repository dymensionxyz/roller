package consts

var DaAuthTokenType = struct {
	Admin string
	Read  string
}{
	Admin: "admin",
	Read:  "read",
}

const (
	DefaultCelestiaRestApiEndpoint = "https://api.celestia-mocha.com"
	DefaultCelestiaRPC             = "http://mocha-4-consensus.mesa.newmetric.xyz:26657"

	// https://docs.celestia.org/nodes/mocha-testnet#community-data-availability-da-grpc-endpoints-for-state-access
	DefaultCelestiaNetwork = "mocha-4"
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
		ApiUrl:           DefaultCelestiaRestApiEndpoint,
		ID:               CelestiaTestnet,
		RpcUrl:           DefaultCelestiaRPC,
		CurrentStateNode: "mocha-4-consensus.mesa.newmetric.xyz",
		StateNodes: []string{
			"mocha-4-consensus.mesa.newmetric.xyz",
			"public-celestia-mocha4-consensus.numia.xyz",
			"mocha-4-consensus.mesa.newmetric.xyz",
			"full.consensus.mocha-4.celestia-mocha.com",
			"consensus-full-mocha-4.celestia-mocha.com",
			"rpc-mocha.pops.one",
		},
		GasPrice: "0.02",
	},
	string(CelestiaMainnet): {
		Backend:          Celestia,
		ApiUrl:           "https://api.celestia.pops.one",
		ID:               CelestiaMainnet,
		RpcUrl:           "http://rpc.celestia.pops.one:26657",
		CurrentStateNode: "rpc.celestia.pops.one",
		StateNodes: []string{
			"rpc-celestia.alphab.ai",
			"celestia.rpc.kjnodes.com",
		},
		GasPrice: "0.002",
	},
}
