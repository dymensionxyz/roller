package consts

var MainnetHubData = HubData{
	ApiUrl:        "https://dymension-mainnet-rest.public.blastapi.io",
	ID:            MainnetHubID,
	RpcUrl:        "https://dymension-mainnet-tendermint.public.blastapi.io",
	ArchiveRpcUrl: "https://dymension-mainnet-tendermint.public.blastapi.io",
	GasPrice:      "20000000000",
	DaNetwork:     CelestiaMainnet,
}

var TestnetHubData = HubData{
	ApiUrl:        "https://api-blumbus.mzonder.com",
	ID:            TestnetHubID,
	RpcUrl:        "https://rpc-blumbus.mzonder.com",
	ArchiveRpcUrl: "https://rpc-blumbus-archive.mzonder.com",
	GasPrice:      "20000000000",
	DaNetwork:     CelestiaTestnet,
}

var DevnetHubData = HubData{
	ApiUrl:        "http://52.58.111.62:1318",
	ID:            DevnetHubID,
	RpcUrl:        "http://52.58.111.62:36657",
	ArchiveRpcUrl: "http://52.58.111.62:36657",
	GasPrice:      "100000000",
	DaNetwork:     CelestiaTestnet,
}

var LocalHubData = HubData{
	ApiUrl:        "http://localhost:1318",
	ID:            LocalHubID,
	RpcUrl:        "http://localhost:36657",
	ArchiveRpcUrl: "http://localhost:36657",
	GasPrice:      "100000000",
	DaNetwork:     MockDA,
}

var MockHubData = HubData{
	ApiUrl:        "",
	ID:            MockHubID,
	RpcUrl:        "",
	ArchiveRpcUrl: "",
	GasPrice:      "",
	DaNetwork:     MockDA,
}

var PlaygroundHubData = HubData{
	ApiUrl:        "https://api-dymension-playground-2.mzonder.com:443",
	ID:            PlaygroundHubID,
	RpcUrl:        "https://rpc-dymension-playground-2.mzonder.com:443",
	ArchiveRpcUrl: "https://rpc-dymension-playground-2.mzonder.com:443",
	GasPrice:      "2000000000",
	DaNetwork:     CelestiaTestnet,
}

// TODO(#112): The available hub networks should be read from YAML file
var Hubs = map[string]HubData{
	MockHubName:       MockHubData,
	LocalHubName:      LocalHubData,
	DevnetHubName:     DevnetHubData,
	TestnetHubName:    TestnetHubData,
	PlaygroundHubName: PlaygroundHubData,
	MainnetHubName:    MainnetHubData,
}

const (
	MockHubName       = "mock"
	LocalHubName      = "local"
	DevnetHubName     = "devnet"
	TestnetHubName    = "testnet"
	PlaygroundHubName = "playground"
	MainnetHubName    = "mainnet"
)

const (
	MockHubID       = "mock"
	LocalHubID      = "dymension_100-1"
	DevnetHubID     = "dymension_100-1"
	PlaygroundHubID = "dymension_2019-1"
	TestnetHubID    = "blumbus_111-1"
	MainnetHubID    = "dymension_1100-1"
)
