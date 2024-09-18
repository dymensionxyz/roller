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
	DefaultCelestiaStateNode = "rpc-mocha.pops.one"
	DefaultCelestiaNetwork   = "mocha-4"
)

type DAType string

const (
	Local    DAType = "mock"
	Celestia DAType = "celestia"
	Avail    DAType = "avail"
)

type DaNetwork string

const (
	CelestiaTestnet DaNetwork = "mocha-4"
	CelestiaMainnet DaNetwork = "celestia"
)

var DaNetworks = map[string]DaData{
	"mock": {
		Backend:   Local,
		ApiUrl:    "",
		ID:        "mock",
		RpcUrl:    "",
		StateNode: "",
		GasPrice:  "",
	},
	"mocha-4": {
		Backend:   Celestia,
		ApiUrl:    DefaultCelestiaRestApiEndpoint,
		ID:        CelestiaTestnet,
		RpcUrl:    DefaultCelestiaRPC,
		StateNode: DefaultCelestiaStateNode,
		GasPrice:  "0.02",
	},
	"celestia": {
		Backend:   Celestia,
		ApiUrl:    "api-celestia.mzonder.com",
		ID:        CelestiaMainnet,
		RpcUrl:    "rpc-celestia.mzonder.com",
		StateNode: "",
		GasPrice:  "0.002",
	},
}
