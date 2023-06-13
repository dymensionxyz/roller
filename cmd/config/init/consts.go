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


var HubData = struct {
	API_URL string
	ID      string
	RPC_URL string
}{
	API_URL: "https://rest-hub-35c.dymension.xyz",
	ID:      "35-C",
	RPC_URL: "https://rpc-hub-35c.dymension.xyz:443",
}

var Executables = struct {
	Celestia  string
	Rollapp   string
	Relayer   string
	Dymension string
}{
	Celestia:  "/usr/local/bin/roller_bins/celestia",
	Rollapp:   "/usr/local/bin/rollapp_evm",
	Relayer:   "/usr/local/bin/roller_bins/rly",
	Dymension: "/usr/local/bin/roller_bins/dymd",
}

const defaultRollappRPC = "http://localhost:26657"

const KeysDirName = "keys"
const RollerConfigFileName = "config.toml"
