package consts

var MainnetHubData = HubData{
	Environment:   "mainnet",
	ApiUrl:        "https://dymension-mainnet-rest.public.blastapi.io:443",
	ID:            MainnetHubID,
	RpcUrl:        "https://dymension-mainnet-tendermint.public.blastapi.io:443",
	ArchiveRpcUrl: "https://dymension-mainnet-tendermint.public.blastapi.io:443",
	GasPrice:      "7000000000",
	DaNetwork:     CelestiaMainnet,
}

var BlumbusHubData = HubData{
	Environment:   "blumbus",
	ApiUrl:        "https://api-blumbus.mzonder.com:443",
	ID:            BlumbusHubID,
	RpcUrl:        "https://rpc-blumbus.mzonder.com:443",
	ArchiveRpcUrl: "https://rpc-blumbus-archive.mzonder.com:443",
	GasPrice:      "20000000000",
	DaNetwork:     CelestiaTestnet,
}

var LocalHubData = HubData{
	Environment:   "local",
	ApiUrl:        "http://localhost:1318",
	ID:            LocalHubID,
	RpcUrl:        "http://localhost:36657",
	ArchiveRpcUrl: "http://localhost:36657",
	GasPrice:      "100000000",
	DaNetwork:     MockDA,
}

var MockHubData = HubData{
	Environment:   "mock",
	ApiUrl:        "",
	ID:            MockHubID,
	RpcUrl:        "",
	ArchiveRpcUrl: "",
	GasPrice:      "",
	DaNetwork:     MockDA,
}

var PlaygroundHubData = HubData{
	Environment:   "playground",
	ApiUrl:        "https://api-dym-migration-test-2.mzonder.com:443",
	ID:            PlaygroundHubID,
	RpcUrl:        "https://rpc-dym-migration-test-2.mzonder.com:443",
	ArchiveRpcUrl: "https://rpc-dym-migration-test-2.mzonder.com:443",
	GasPrice:      "2000000000",
	DaNetwork:     CelestiaTestnet,
}

// TODO(#112): The available hub networks should be read from YAML file
var Hubs = map[string]HubData{
	MockHubName:       MockHubData,
	LocalHubName:      LocalHubData,
	BlumbusHubName:    BlumbusHubData,
	PlaygroundHubName: PlaygroundHubData,
	MainnetHubName:    MainnetHubData,
}

const (
	MockHubName       = "mock"
	LocalHubName      = "local"
	BlumbusHubName    = "blumbus"
	PlaygroundHubName = "playground"
	MainnetHubName    = "mainnet"
)

const (
	MockHubID       = "mock"
	LocalHubID      = "dymension_100-1"
	PlaygroundHubID = "dymension_1405-1"
	BlumbusHubID    = "blumbus_111-1"
	MainnetHubID    = "dymension_1100-1"
)
