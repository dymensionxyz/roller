package consts

import (
	"fmt"
)

const (
	binsDir                  = "/usr/local/bin" // TODO : change me if binaries path was different
	DefaultTokenSupply       = "1000000000000000000000000000"
	DefaultFee               = 2000000000000000000    // 2
	DefaultTxFee             = 10000000000000000      // 0.01
	DefaultAdditionalFunding = "20000000000000000000" // 20, this amount is added to the necessary sequencer balance
)

var InternalBinsDir = fmt.Sprintf("%s/roller_bins", binsDir)

var (
	AllServices            = []string{"rollapp", "da-light-client", "relayer", "eibc"}
	RollappSystemdServices = []string{"rollapp", "da-light-client"}
	RelayerSystemdServices = []string{"relayer"}
	EibcSystemdServices    = []string{"eibc"}
)

var Executables = struct {
	Celestia    string
	RollappEVM  string
	Relayer     string
	Dymension   string
	CelKey      string
	Roller      string
	Simd        string
	Eibc        string
	CelestiaApp string
}{
	Roller:      fmt.Sprintf("%s/roller", binsDir),
	RollappEVM:  fmt.Sprintf("%s/rollapp-evm", binsDir), // changed rollappd to rollapp-evm binary
	Dymension:   fmt.Sprintf("%s/dymd", binsDir),
	Celestia:    fmt.Sprintf("%s/celestia", InternalBinsDir),
	CelKey:      fmt.Sprintf("%s/cel-key", InternalBinsDir),
	Relayer:     fmt.Sprintf("%s/rly", InternalBinsDir),
	Simd:        fmt.Sprintf("%s/simd", InternalBinsDir),
	Eibc:        fmt.Sprintf("%s/eibc-client", binsDir),
	CelestiaApp: fmt.Sprintf("%s/celestia-appd", InternalBinsDir),
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
}{
	HubSequencer:                  "hub_sequencer",
	HubGenesis:                    "hub_genesis",
	RollappSequencer:              "rollapp_genesis_account",
	RollappSequencerReward:        "rollapp_sequencer_rewards",
	RollappSequencerPrivValidator: "rollapp_sequencer_priv_validator",
	RollappRelayer:                "relayer-rollapp-key",
	HubRelayer:                    "relayer-hub-key",
	Celestia:                      "", // my_celes_key
	Eibc:                          "whale",
	Da:                            "da_key",
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
}{
	Rollapp:              "rollapp",
	Relayer:              "relayer",
	DALightNode:          "da-light-node",
	HubKeys:              "hub-keys",
	RollappSequencerKeys: "rollapp-sequencer-keys",
	LocalHub:             "local-hub",
	Eibc:                 ".eibc-client",
	BlockExplorer:        "block-explorer",
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

var NodeType = struct {
	Sequencer string
	FullNode  string
}{
	Sequencer: "sequencer",
	FullNode:  "fullnode",
}

const RollerConfigFileName = "roller.toml"

type VMType string

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
