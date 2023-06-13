package consts

var Executables = struct {
	Celestia   string
	RollappEVM string
	Relayer    string
	Dymension  string
}{
	Celestia:   "/usr/local/bin/roller_bins/celestia",
	RollappEVM: "/usr/local/bin/rollapp_evm",
	Relayer:    "/usr/local/bin/roller_bins/rly",
	Dymension:  "/usr/local/bin/roller_bins/dymd",
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
	DALightNode:      "my-celes-key",
	HubRelayer:       "relayer-hub-key",
}
