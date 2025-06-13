package roller

import "github.com/dymensionxyz/roller/cmd/consts"

var SupportedDas = []consts.DAType{consts.Celestia, consts.Avail, consts.LoadNetwork, consts.Bnb, consts.Aptos, consts.Sui, consts.Walrus, consts.Ethereum, consts.Local}

type RollappConfig struct {
	// new roller.toml
	Home           string                         `toml:"home"`
	RollerVersion  string                         `toml:"roller_version"`
	KeyringBackend consts.SupportedKeyringBackend `toml:"keyring_backend"`

	NodeType string `toml:"node_type"`

	GenesisHash string `toml:"genesis_hash"`
	GenesisUrl  string `toml:"genesis_url"`
	RollappID   string `toml:"rollapp_id"`

	Environment string `toml:"environment"`

	RollappVMType        consts.VMType `toml:"rollapp_vm_type"`
	RollappBinary        string        `toml:"rollapp_binary"`
	RollappBinaryVersion string        `toml:"rollapp_binary_version"`
	Bech32Prefix         string        `toml:"bech32_prefix"`
	BaseDenom            string        `toml:"base_denom"`
	Denom                string        `toml:"denom"`
	Decimals             uint
	MinGasPrices         string `toml:"minimum_gas_prices"`

	HubData     consts.HubData    `toml:"HubData"`
	DA          consts.DaData     `toml:"DA"`
	HealthAgent HealthAgentConfig `toml:"HealthAgent"`
}

type HealthAgentConfig struct {
	Enabled bool `toml:"enabled"`
}
