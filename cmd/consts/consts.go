package consts

import "fmt"

const binsDir = "/usr/local/bin"

var internalBinsDir = fmt.Sprintf("%s/roller_bins", binsDir)

var Executables = struct {
	Celestia   string
	RollappEVM string
	Relayer    string
	Dymension  string
	CelKey     string
}{
	Celestia:   fmt.Sprintf("%s/celestia", internalBinsDir),
	CelKey:     fmt.Sprintf("%s/cel-key", internalBinsDir),
	RollappEVM: fmt.Sprintf("%s/rollapp_evm", binsDir),
	Relayer:    fmt.Sprintf("%s/rly", internalBinsDir),
	Dymension:  fmt.Sprintf("%s/dymd", internalBinsDir),
}

var KeyNames = struct {
	HubSequencer     string
	RollappSequencer string
	RollappRelayer   string
	DALightNode      string
	HubRelayer       string
}{
	HubSequencer:     "hub_sequencer",
	RollappSequencer: "rollapp_sequencer",
	RollappRelayer:   "relayer-rollapp-key",
	DALightNode:      "my_celes_key",
	HubRelayer:       "relayer-hub-key",
}
