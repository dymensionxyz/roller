package consts

var MainnetHubData = HubData{
	API_URL:         "https://dymension-mainnet-rest.public.blastapi.io",
	ID:              MainnetHubID,
	RPC_URL:         "https://dymension-mainnet-tendermint.public.blastapi.io",
	ARCHIVE_RPC_URL: "https://dymension-mainnet-tendermint.public.blastapi.io",
	GAS_PRICE:       "20000000000",
}

var TestnetHubData = HubData{
	API_URL:         "https://api-blumbus.mzonder.com",
	ID:              TestnetHubID,
	RPC_URL:         "https://rpc-blumbus.mzonder.com",
	ARCHIVE_RPC_URL: "https://rpc-blumbus-archive.mzonder.com",
	GAS_PRICE:       "20000000000",
}

var DevnetHubData = HubData{
	API_URL:         "http://52.58.111.62:1318",
	ID:              DevnetHubID,
	RPC_URL:         "http://52.58.111.62:36657",
	ARCHIVE_RPC_URL: "http://52.58.111.62:36657",
	GAS_PRICE:       "100000000",
}

var LocalHubData = HubData{
	API_URL:         "http://localhost:1318",
	ID:              LocalHubID,
	RPC_URL:         "http://localhost:36657",
	ARCHIVE_RPC_URL: "http://localhost:36657",
	GAS_PRICE:       "100000000",
}

var MockHubData = HubData{
	API_URL:         "",
	ID:              MockHubID,
	RPC_URL:         "",
	ARCHIVE_RPC_URL: "",
	GAS_PRICE:       "",
}

// TODO(#112): The available hub networks should be read from YAML file
var Hubs = map[string]HubData{
	MockHubName:    MockHubData,
	LocalHubName:   LocalHubData,
	DevnetHubName:  DevnetHubData,
	TestnetHubName: TestnetHubData,
	MainnetHubName: MainnetHubData,
}

const (
	MockHubName    = "mock"
	LocalHubName   = "local"
	DevnetHubName  = "devnet"
	TestnetHubName = "testnet"
	MainnetHubName = "mainnet"
)

const (
	MockHubID    = "mock"
	LocalHubID   = "dymension_100-1"
	DevnetHubID  = "dymension_100-1"
	TestnetHubID = "blumbus_111-1"
	MainnetHubID = "dymension_1100-1"
)
