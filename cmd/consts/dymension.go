package consts

var MainnetHubData = HubData{
	Environment:   "mainnet",
	ApiUrl:        "https://dymension-mainnet-rest.public.blastapi.io:443",
	ID:            MainnetHubID,
	RpcUrl:        "https://dymension-mainnet-tendermint.public.blastapi.io:443",
	WsUrl:         "",
	ArchiveRpcUrl: "https://dymension-mainnet-tendermint.public.blastapi.io:443",
	GasPrice:      "7000000000",
}

var BlumbusHubData = HubData{
	Environment:   "blumbus",
	ApiUrl:        "https://api-blumbus.mzonder.com:443",
	ID:            BlumbusHubID,
	RpcUrl:        "https://rpc-blumbus.mzonder.com:443",
	WsUrl:         "https://rpc-blumbus.mzonder.com:443",
	ArchiveRpcUrl: "https://rpc-blumbus-archive.mzonder.com:443",
	GasPrice:      "20000000000",
}

var LocalHubData = HubData{
	Environment:   "local",
	ApiUrl:        "http://localhost:1318",
	ID:            LocalHubID,
	RpcUrl:        "http://localhost:36657",
	WsUrl:         "http://localhost:36657",
	ArchiveRpcUrl: "http://localhost:36657",
	GasPrice:      "100000000",
}

var MockHubData = HubData{
	Environment:   "mock",
	ApiUrl:        "",
	ID:            MockHubID,
	RpcUrl:        "",
	WsUrl:         "",
	ArchiveRpcUrl: "",
	GasPrice:      "",
}

var PlaygroundHubData = HubData{
	Environment:   "playground",
	ApiUrl:        "https://api-dymension-playground35.mzonder.com:443",
	ID:            PlaygroundHubID,
	RpcUrl:        "https://rpc-dymension-playground35.mzonder.com:443",
	WsUrl:         "https://rpc-dymension-playground35.mzonder.com:443",
	ArchiveRpcUrl: "https://rpc-dymension-playground35.mzonder.com:443",
	GasPrice:      "2000000000",
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
	PlaygroundHubID = "dymension_3405-1"
	BlumbusHubID    = "blumbus_111-1"
	MainnetHubID    = "dymension_1100-1"
)
