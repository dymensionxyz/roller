package consts

var MainnetHubData = HubData{
	Environment:   "mainnet",
	ApiUrl:        "https://dymension-mainnet-rest.public.blastapi.io",
	ID:            MainnetHubID,
	RpcUrl:        "https://dymension-mainnet-tendermint.public.blastapi.io",
	ArchiveRpcUrl: "https://dymension-mainnet-tendermint.public.blastapi.io",
	GasPrice:      "20000000000",
	// DaNetwork:      string(CelestiaMainnet),
	DaNetwork: AvailMainnet,
}

var TestnetHubData = HubData{
	Environment:   "blumbus",
	ApiUrl:        "https://api-blumbus.mzonder.com",
	ID:            TestnetHubID,
	RpcUrl:        "https://rpc-blumbus.mzonder.com",
	ArchiveRpcUrl: "https://rpc-blumbus-archive.mzonder.com",
	GasPrice:      "20000000000",
	DaNetwork:     AvailTestnet,
	// DaNetwork:      string(CelestiaTestnet),
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
	ApiUrl:        "http://localhost:1318",
	ID:            PlaygroundHubID,
	RpcUrl:        "http://localhost:36657",
	ArchiveRpcUrl: "http://localhost:36657",
	GasPrice:      "2000000000",
	DaNetwork:     AvailTestnet,
	// ApiUrl:        "https://api-dym-migration-test-2.mzonder.com:443",
	// RpcUrl:        "https://rpc-dym-migration-test-2.mzonder.com:443",
	// ArchiveRpcUrl: "https://rpc-dym-migration-test-2.mzonder.com:443",
}

// TODO(#112): The available hub networks should be read from YAML file
var Hubs = map[string]HubData{
	MockHubName:       MockHubData,
	LocalHubName:      LocalHubData,
	TestnetHubName:    TestnetHubData,
	PlaygroundHubName: PlaygroundHubData,
	MainnetHubName:    MainnetHubData,
}

const (
	MockHubName       = "mock"
	LocalHubName      = "local"
	TestnetHubName    = "testnet"
	PlaygroundHubName = "playground"
	MainnetHubName    = "mainnet"
)

const (
	MockHubID  = "mock"
	LocalHubID = "dymension_100-1"
	// PlaygroundHubID = "dymension_1405-1"
	PlaygroundHubID = "dymension_100-1"
	TestnetHubID    = "blumbus_111-1"
	MainnetHubID    = "dymension_1100-1"
)
