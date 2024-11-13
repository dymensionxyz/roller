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
	AvailTestnet    DaNetwork = "avail"
	AvailMainnet    DaNetwork = "avail-1" // change this with correct mainnet id
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
		ApiUrl:           "api-celestia.mzonder.com",
		ID:               CelestiaMainnet,
		RpcUrl:           "rpc-celestia.mzonder.com",
		CurrentStateNode: "",
		StateNodes: []string{
			"",
		},
		GasPrice: "0.002",
	},
	string(AvailTestnet): {
		Backend:          Avail,
		ApiUrl:           "http://localhost:8000",
		ID:               AvailTestnet,
		RpcUrl:           "ws://127.0.0.1:9944",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(AvailMainnet): {
		Backend:          Avail,
		ApiUrl:           "http://localhost:8000",
		ID:               AvailMainnet,
		RpcUrl:           "ws://127.0.0.1:9944",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
}
