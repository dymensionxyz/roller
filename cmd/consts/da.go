package consts

var DaAuthTokenType = struct {
	Admin string
	Read  string
}{
	Admin: "admin",
	Read:  "read",
}

type DAType string

const (
	Local    DAType = "local"
	Celestia DAType = "celestia"
	Avail    DAType = "avail"
)

var CelestiaNetworks = map[string]HubData{
	"mocha": {
		API_URL:         "celestia-testnet-consensus.itrocket.net",
		ID:              "mocha-4",
		RPC_URL:         "celestia-testnet-consensus.itrocket.net",
		ARCHIVE_RPC_URL: "",
		GAS_PRICE:       "",
	},
	"celestia": {
		API_URL:         "api-celestia.mzonder.com",
		ID:              "celestia",
		RPC_URL:         "rpc-celestia.mzonder.com",
		ARCHIVE_RPC_URL: "",
		GAS_PRICE:       "",
	},
}
