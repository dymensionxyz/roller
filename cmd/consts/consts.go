package consts

import (
	"fmt"
)

const (
	binsDir              = "/usr/local/bin"
	DefaultTokenSupply   = "1000000000000000000000000000"
	DefaultFee           = 2000000000000000000   // 2
	DefaultTxFee         = 10000000000000000     // 0.01
	MinOperationalAmount = "5000000000000000000" // 5, is checked on every rollapp process interaction ( start, restart )
	DefaultMinGasPrice   = "2000000000"
)

var InternalBinsDir = fmt.Sprintf("%s/roller_bins", binsDir)

var (
	AllServices                     = []string{"rollapp", "da-light-client", "relayer", "eibc"}
	RollappWithCelesSystemdServices = []string{"rollapp", "da-light-client"}
	RollappSystemdServices          = []string{"rollapp"}
	RelayerSystemdServices          = []string{"relayer"}
	PriceOracleSystemdServices      = []string{"price"}
	RngOracleSystemdServices        = []string{"rng"}
	EibcSystemdServices             = []string{"eibc"}
	AlertAgentSystemdServices       = []string{"alert-agent"}
)

var Executables = struct {
	Celestia               string
	RollappEVM             string
	Relayer                string
	Dymension              string
	CelKey                 string
	Roller                 string
	Simd                   string
	Eibc                   string
	CelestiaApp            string
	PriceOracle            string
	RngOracle              string
	RngOracleRandomService string
	Solc                   string
	AlertAgent             string
}{
	Roller:      fmt.Sprintf("%s/roller", binsDir),
	RollappEVM:  fmt.Sprintf("%s/rollappd", binsDir),
	Dymension:   fmt.Sprintf("%s/dymd", binsDir),
	Celestia:    fmt.Sprintf("%s/celestia", InternalBinsDir),
	CelKey:      fmt.Sprintf("%s/cel-key", InternalBinsDir),
	Relayer:     fmt.Sprintf("%s/rly", InternalBinsDir),
	Simd:        fmt.Sprintf("%s/simd", InternalBinsDir),
	Eibc:        fmt.Sprintf("%s/eibc-client", binsDir),
	CelestiaApp: fmt.Sprintf("%s/celestia-appd", InternalBinsDir),
	// oracles
	PriceOracle:            fmt.Sprintf("%s/price-oracle", InternalBinsDir),
	RngOracle:              fmt.Sprintf("%s/rng-oracle", InternalBinsDir),
	RngOracleRandomService: fmt.Sprintf("%s/rng-oracle-random-service", InternalBinsDir),
	Solc:                   fmt.Sprintf("%s/solc", InternalBinsDir),
	AlertAgent:             fmt.Sprintf("%s/alert-agent", InternalBinsDir),
}

var KeysIds = struct {
	HubSequencer                  string
	HubGenesis                    string
	RollappSequencer              string
	RollappSequencerReward        string
	RollappSequencerPrivValidator string
	RollappRelayer                string
	HubRelayer                    string
	Celestia                      string
	Eibc                          string
	Da                            string
	Oracle                        string
}{
	HubSequencer:                  "hub_sequencer",
	HubGenesis:                    "hub_genesis",
	RollappSequencer:              "rollapp_genesis_account",
	RollappSequencerReward:        "rollapp_sequencer_rewards",
	RollappSequencerPrivValidator: "rollapp_sequencer_priv_validator",
	RollappRelayer:                "relayer-rollapp-key",
	HubRelayer:                    "relayer-hub-key",
	Celestia:                      "my_celes_key",
	Eibc:                          "whale",
	Da:                            "da_key",
	Oracle:                        "oracle",
}

var Oracles = struct {
	Rng   string
	Price string
}{
	Rng:   "rng",
	Price: "price",
}

var AddressPrefixes = struct {
	Hub string
}{
	Hub: "dym",
}

var ConfigDirName = struct {
	Rollapp              string
	Relayer              string
	DALightNode          string
	HubKeys              string
	RollappSequencerKeys string
	LocalHub             string
	Eibc                 string
	BlockExplorer        string
	Oracle               string
	Snapshots            string
	AlertAgent           string
}{
	Rollapp:              "rollapp",
	Relayer:              "relayer",
	DALightNode:          "da-light-node",
	HubKeys:              "hub-keys",
	RollappSequencerKeys: "rollapp-sequencer-keys",
	LocalHub:             "local-hub",
	Eibc:                 ".eibc-client",
	BlockExplorer:        "block-explorer",
	Oracle:               "oracle",
	Snapshots:            "snapshots",
	AlertAgent:           "alert-agent",
}

var Denoms = struct {
	Hub             string
	HubIbcOnRollapp string
	Celestia        string
	Avail           string
	LoadNetwork     string
	Bnb             string
}{
	Hub:             "adym",
	HubIbcOnRollapp: "ibc/FECACB927EB3102CCCB240FFB3B6FCCEEB8D944C6FEA8DFF079650FEFF59781D",
	Celestia:        "utia",
	Avail:           "aAVL",
	LoadNetwork:     "tLOAD",
	Bnb:             "ubnb",
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

var NodeType = struct {
	Sequencer string
	FullNode  string
}{
	Sequencer: "sequencer",
	FullNode:  "fullnode",
}

const RollerConfigFileName = "roller.toml"

type VMType string

func (v VMType) String() string {
	return string(v)
}

const (
	SDK_ROLLAPP  VMType = "sdk"
	EVM_ROLLAPP  VMType = "evm"
	WASM_ROLLAPP VMType = "wasm"
)

func ToVMType(s string) (VMType, error) {
	switch s {
	case string(SDK_ROLLAPP):
		return SDK_ROLLAPP, nil
	case string(EVM_ROLLAPP):
		return EVM_ROLLAPP, nil
	case string(WASM_ROLLAPP):
		return WASM_ROLLAPP, nil
	default:
		return "", fmt.Errorf("invalid VMType: %s", s)
	}
}
