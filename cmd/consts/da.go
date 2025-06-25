package consts

var DaAuthTokenType = struct {
	Admin string
	Read  string
}{
	Admin: "admin",
	Read:  "read",
}

const (
	DefaultCelestiaMochaRest    = "https://api.celestia-mocha.com"
	DefaultCelestiaMochaRPC     = "https://celestia-testnet-rpc.itrocket.net:443"
	DefaultCelestiaMochaNetwork = "mocha-4"
	// https://docs.celestia.org/nodes/mocha-testnet#community-data-availability-da-grpc-endpoints-for-state-access
	DefaultCelestiaRest    = "https://api.celestia.pops.one"
	DefaultCelestiaRPC     = "http://rpc.celestia.pops.one:26657"
	DefaultCelestiaNetwork = "celestia"
)

type DAType string

const (
	Local       DAType = "mock"
	Celestia    DAType = "celestia"
	Avail       DAType = "avail"
	Aptos       DAType = "aptos"
	LoadNetwork DAType = "loadnetwork"
	Bnb         DAType = "bnb"
	Sui         DAType = "sui"
	Walrus      DAType = "walrus"
	Mock        DAType = "mock"
	Solana      DAType = "solana"
	Ethereum    DAType = "ethereum"
	Kaspa       DAType = "kaspa"
)

type DaNetwork string

const (
	MockDA             DaNetwork = "mock"
	CelestiaTestnet    DaNetwork = "mocha-4"
	CelestiaMainnet    DaNetwork = "celestia"
	AvailTestnet       DaNetwork = "avail"
	AvailMainnet       DaNetwork = "avail-1" // change this with correct mainnet id
	LoadNetworkTestnet DaNetwork = "alphanet"
	LoadNetworkMainnet DaNetwork = "loadnetwork" // change this with correct mainnet id
	BnbTestnet         DaNetwork = "97"
	BnbMainnet         DaNetwork = "56" // change this with correct mainnet id
	SuiTestnet         DaNetwork = "sui-testnet"
	SuiMainnet         DaNetwork = "sui-mainnet"
	// https://aptos.dev/en/network/nodes/networks
	AptosTestnet  DaNetwork = "2"
	AptosMainnet  DaNetwork = "1"
	WalrusTestnet DaNetwork = "walrus-testnet"
	WalrusMainnet DaNetwork = "walrus-mainnet" // change this with correct mainnet id
	SolanaTestnet DaNetwork = "solana-testnet"
	SolanaMainnet DaNetwork = "solana-mainnet"
	EthereumTestnet DaNetwork = "eth-testnet"
	EthereumMainnet DaNetwork = "eth-mainnet"
	KaspaTestnet  DaNetwork = "kaspa-testnet"
	KaspaMainnet  DaNetwork = "kaspa-mainnet"
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
		ApiUrl:           DefaultCelestiaMochaRest,
		ID:               CelestiaTestnet,
		RpcUrl:           DefaultCelestiaMochaRPC,
		CurrentStateNode: "rpc-mocha.pops.one",
		StateNodes: []string{
			"public-celestia-mocha4-consensus.numia.xyz",
			"full.consensus.mocha-4.celestia-mocha.com",
			"consensus-full-mocha-4.celestia-mocha.com",
			"rpc-mocha.pops.one",
		},
		GasPrice: "0.02",
	},
	string(CelestiaMainnet): {
		Backend:          Celestia,
		ApiUrl:           DefaultCelestiaRest,
		ID:               CelestiaMainnet,
		RpcUrl:           DefaultCelestiaRPC,
		CurrentStateNode: "rpc.celestia.pops.one",
		StateNodes: []string{
			"rpc-celestia.alphab.ai",
			"celestia.rpc.kjnodes.com",
		},
		GasPrice: "0.002",
	},
	string(AvailTestnet): {
		Backend:          Avail,
		ApiUrl:           "https://turing-rpc.avail.so/rpc",
		ID:               AvailTestnet,
		RpcUrl:           "wss://turing-rpc.avail.so/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(AvailMainnet): {
		Backend:          Avail,
		ApiUrl:           "https://mainnet-rpc.avail.so/rpc:443",
		ID:               AvailMainnet,
		RpcUrl:           "wss://mainnet.avail-rpc.com/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(LoadNetworkTestnet): {
		Backend:          LoadNetwork,
		ApiUrl:           "https://alphanet.load.network",
		ID:               LoadNetworkTestnet,
		RpcUrl:           "wss://alphanet.load.network/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(LoadNetworkMainnet): {
		Backend:          LoadNetwork,
		ApiUrl:           "",
		ID:               LoadNetworkMainnet,
		RpcUrl:           "",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(BnbTestnet): {
		Backend:          Bnb,
		ApiUrl:           "https://data-seed-prebsc-1-s1.bnbchain.org:8545",
		ID:               BnbTestnet,
		RpcUrl:           "https://data-seed-prebsc-1-s1.bnbchain.org:8545",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(BnbMainnet): {
		Backend:          Bnb,
		ApiUrl:           "https://bsc-dataseed.bnbchain.org",
		ID:               BnbMainnet,
		RpcUrl:           "https://bsc-dataseed.bnbchain.org",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(AptosTestnet): {
		Backend:          Aptos,
		ApiUrl:           "",
		ID:               AptosTestnet,
		RpcUrl:           "",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(AptosMainnet): {
		Backend:          Aptos,
		ApiUrl:           "",
		ID:               AptosMainnet,
		RpcUrl:           "",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(SuiTestnet): {
		Backend:          Sui,
		ApiUrl:           "",
		ID:               SuiTestnet,
		RpcUrl:           "https://fullnode.testnet.sui.io:443",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(SuiMainnet): {
		Backend:          Sui,
		ApiUrl:           "",
		ID:               SuiMainnet,
		RpcUrl:           "https://fullnode.mainnet.sui.io:443",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(WalrusTestnet): {
		Backend:          Walrus,
		ApiUrl:           "https://aggregator.walrus-testnet.walrus.space",
		ID:               WalrusTestnet,
		RpcUrl:           "https://publisher.walrus-testnet.walrus.space",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(WalrusMainnet): {
		Backend:          Walrus,
		ApiUrl:           "https://aggregator.walrus-mainnet.walrus.space",
		ID:               WalrusMainnet,
		RpcUrl:           "https://publisher.walrus-mainnet.walrus.space",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(SolanaTestnet): {
		Backend:          Solana,
		ApiUrl:           "http://barcelona:8899",
		ID:               SolanaTestnet,
		RpcUrl:           "http://barcelona:8899",
    CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "0.00000002",
  },
	string(SolanaMainnet): {
		Backend:          Solana,
		ApiUrl:           "http://barcelona:8899",
		ID:               SolanaMainnet,
		RpcUrl:           "http://barcelona:8899",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(EthereumTestnet): {
		Backend:          Ethereum,
		ApiUrl:           "https://ethereum-sepolia-beacon-api.publicnode.com",
		ID:               EthereumTestnet,
		RpcUrl:           "https://ethereum-sepolia-rpc.publicnode.com",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "0.00000002",
	},
	string(EthereumMainnet): {
		Backend:          Ethereum,
		ApiUrl:           "https://ethereum-beacon-api.publicnode.com",
		ID:               EthereumMainnet,
		RpcUrl:           "https://ethereum-rpc.publicnode.com",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "0.00000002",
	string(KaspaTestnet): {
		Backend:          Kaspa,
		ApiUrl:           "https://api-tn10.kaspa.org",
		ID:               KaspaTestnet,
		RpcUrl:           "wss://testnet.kaspa.org/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
	string(KaspaMainnet): {
		Backend:          Kaspa,
		ApiUrl:           "https://api.kaspa.org",
		ID:               KaspaMainnet,
		RpcUrl:           "wss://mainnet.kaspa.org/ws",
		CurrentStateNode: "",
		StateNodes:       []string{},
		GasPrice:         "",
	},
}
