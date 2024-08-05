package dymint

import (
	"fmt"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
	"github.com/pterm/pterm"
)

// TODO: use dymint instead
type dymintConfig struct {
	BatchAcceptanceAttempts    string                      `toml:"batch_acceptance_attempts"`
	BatchAcceptanceTimeout     string                      `toml:"batch_acceptance_timeout"`
	BatchSubmitMaxTime         string                      `toml:"batch_submit_max_time"`
	BlockBatchMaxSizeBytes     int                         `toml:"block_batch_max_size_bytes"`
	BlockBatchSize             string                      `toml:"block_batch_size"`
	BlockTime                  string                      `toml:"block_time"`
	DaConfig                   string                      `toml:"da_config"`
	DaLayer                    string                      `toml:"da_layer"`
	DymAccountName             string                      `toml:"dym_account_name"`
	EmptyBlocksMaxTime         string                      `toml:"empty_blocks_max_time"`
	GasPrices                  string                      `toml:"gas_prices"`
	KeyringBackend             string                      `toml:"keyring_backend"`
	KeyringHomeDir             string                      `toml:"keyring_home_dir"`
	MaxIdleTime                string                      `toml:"max_idle_time"`
	MaxProofTime               string                      `toml:"max_proof_time"`
	MaxSupportedBatchSkew      int                         `toml:"max_supported_batch_skew"`
	NamespaceID                string                      `toml:"namespace_id"`
	NodeAddress                string                      `toml:"node_address"`
	P2PAdvertisingEnabled      string                      `toml:"p2p_advertising_enabled"`
	P2PBootstrapNodes          string                      `toml:"p2p_bootstrap_nodes"`
	P2PBootstrapRetryTime      string                      `toml:"p2p_bootstrap_retry_time"`
	P2PGossipedBlocksCacheSize int                         `toml:"p2p_gossiped_blocks_cache_size"`
	P2PListenAddress           string                      `toml:"p2p_listen_address"`
	RetryAttempts              string                      `toml:"retry_attempts"`
	RetryMaxDelay              string                      `toml:"retry_max_delay"`
	RetryMinDelay              string                      `toml:"retry_min_delay"`
	RollappID                  string                      `toml:"rollapp_id"`
	SettlementGasFees          string                      `toml:"settlement_gas_fees"`
	SettlementGasLimit         int                         `toml:"settlement_gas_limit"`
	SettlementGasPrices        string                      `toml:"settlement_gas_prices"`
	SettlementLayer            string                      `toml:"settlement_layer"`
	SettlementNodeAddress      string                      `toml:"settlement_node_address"`
	Db                         dymintDBConfig              `toml:"db"`
	Instrumentation            dymintInstrumentationConfig `toml:"instrumentation"`
}

type dymintDBConfig struct {
	InMemory   bool `toml:"in_memory"`
	SyncWrites bool `toml:"sync_writes"`
}

type dymintInstrumentationConfig struct {
	Prometheus           bool   `toml:"prometheus"`
	PrometheusListenAddr string `toml:"prometheus_listen_addr"`
}

func UpdateDymintConfigForIBC(home string) error {
	pterm.Info.Println("checking dymint block time settings")
	dymintPath := sequencer.GetDymintFilePath(home)
	fmt.Println(dymintPath)
	dymintCfg, err := config.LoadConfigFromTOML(dymintPath)
	if err != nil {
		return err
	}

	var cfg dymintConfig

	_, err = toml.Decode(string(dymintCfg), &cfg)
	if err != nil {
		return err
	}

	want := time.Second * 5
	have, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return err
	}

	if want < have {
		pterm.Info.Println(
			"block time is higher then recommended when creating ibc channels: ",
			have,
		)
		pterm.Info.Println("updating dymint config")

		err = utils.UpdateFieldInToml(dymintPath, "max_idle_time", want.String())
		if err != nil {
			return err
		}
		err = utils.UpdateFieldInToml(dymintPath, "batch_submit_max_time", want.String())
		if err != nil {
			return err
		}
		err = utils.UpdateFieldInToml(dymintPath, "max_proof_time", want.String())
		if err != nil {
			return err
		}
	}

	pterm.DefaultInteractiveConfirm.WithDefaultText(
		"would you like roller to restart your rollapp process?",
	)

	return nil
}
