package consts

import (
	"fmt"
)

const (
	binsDir            = "/usr/local/bin"
	DefaultTokenSupply = "1000000000000000000000000000"
	DefaultFee         = 100000000000000000 // 0.1
)

var internalBinsDir = fmt.Sprintf("%s/roller_bins", binsDir)

var Executables = struct {
	Celestia   string
	RollappEVM string
	Relayer    string
	Dymension  string
	CelKey     string
	Roller     string
	Simd       string
	Eibc       string
}{
	Roller:     fmt.Sprintf("%s/roller", binsDir),
	RollappEVM: fmt.Sprintf("%s/rollapp-evm", binsDir),
	Dymension:  fmt.Sprintf("%s/dymd", binsDir),
	Celestia:   fmt.Sprintf("%s/celestia", internalBinsDir),
	CelKey:     fmt.Sprintf("%s/cel-key", internalBinsDir),
	Relayer:    fmt.Sprintf("%s/rly", internalBinsDir),
	Simd:       fmt.Sprintf("%s/simd", internalBinsDir),
	Eibc:       fmt.Sprintf("%s/eibc", binsDir),
}

var KeysIds = struct {
	HubSequencer     string
	HubGenesis       string
	RollappSequencer string
	RollappRelayer   string
	HubRelayer       string
	Celestia         string
}{
	HubSequencer:     "hub_sequencer",
	HubGenesis:       "hub_genesis",
	RollappSequencer: "rollapp_genesis_account",
	RollappRelayer:   "relayer-rollapp-key",
	HubRelayer:       "relayer-hub-key",
	Celestia:         "my_celes_key",
}

var AddressPrefixes = struct {
	Hub string
}{
	Hub: "dym",
}

var ConfigDirName = struct {
	Rollapp     string
	Relayer     string
	DALightNode string
	HubKeys     string
	LocalHub    string
	Eibc        string
}{
	Rollapp:     "rollapp",
	Relayer:     "relayer",
	DALightNode: "da-light-node",
	HubKeys:     "hub-keys",
	LocalHub:    "local-hub",
	Eibc:        ".order-client",
}

var Denoms = struct {
	Hub      string
	Celestia string
	Avail    string
}{
	Hub:      "adym",
	Celestia: "utia",
	Avail:    "aAVL",
}

const (
	KeysDirName        = "keys"
	DefaultRelayerPath = "hub-rollapp"
	DefaultRollappRPC  = "http://localhost:26657"
)

var SpinnerMsgs = struct {
	UniqueIdVerification string
	BalancesVerification string
}{
	UniqueIdVerification: " Verifying unique RollApp ID...\n",
	BalancesVerification: " Verifying balances...\n",
}

type HubData = struct {
	API_URL         string `toml:"api_url"`
	ID              string `toml:"id"`
	RPC_URL         string `toml:"rpc_url"`
	ARCHIVE_RPC_URL string `toml:"archive_rpc_url"`
	GAS_PRICE       string `toml:"gas_price"`
}

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

var NodeType = struct {
	Sequencer string
	FullNode  string
}{
	Sequencer: "sequencer",
	FullNode:  "fullnode",
}

var DaAuthTokenType = struct {
	Admin string
	Read  string
}{
	Admin: "admin",
	Read:  "read",
}

const RollerConfigFileName = "roller.toml"

type VMType string

const (
	SDK_ROLLAPP VMType = "sdk"
	EVM_ROLLAPP VMType = "evm"
)

type DAType string

const (
	Local    DAType = "local"
	Celestia DAType = "celestia"
	Avail    DAType = "avail"
)
