package dymint

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
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

type RollappHealthResponse struct {
	JSONRPC string `json:"jsonrpc"`
	Result  struct {
		IsHealthy bool   `json:"isHealthy"`
		Error     string `json:"error"`
	} `json:"result"`
	ID int `json:"id"`
}

func UpdateDymintConfigForIBC(home string, t string, forceUpdate bool) error {
	pterm.Info.Println("checking dymint block time settings")
	dymintPath := sequencer.GetDymintFilePath(home)
	dymintCfg, err := tomlconfig.Load(dymintPath)
	if err != nil {
		return err
	}

	var cfg dymintConfig

	_, err = toml.Decode(string(dymintCfg), &cfg)
	if err != nil {
		return err
	}

	want, err := time.ParseDuration(t)
	if err != nil {
		return err
	}

	have, err := time.ParseDuration(cfg.MaxIdleTime)
	if err != nil {
		return err
	}

	if want < have || forceUpdate {
		if want < have {
			pterm.Info.Println(
				"block time is higher then recommended when creating ibc channels: ",
				have,
			)
		}
		pterm.Info.Println("updating dymint config")

		err = utils.UpdateFieldInToml(dymintPath, "max_idle_time", want.String())
		if err != nil {
			return err
		}
		err = utils.UpdateFieldInToml(dymintPath, "batch_submit_max_time", want.String())
		if err != nil {
			return err
		}
		err = utils.UpdateFieldInToml(dymintPath, "batch_submit_time", want.String())
		if err != nil {
			return err
		}

		if want < time.Minute*1 {
			err = utils.UpdateFieldInToml(dymintPath, "max_proof_time", want.String())
			if err != nil {
				return err
			}
		} else {
			err = utils.UpdateFieldInToml(dymintPath, "max_proof_time", "1m40s")
			if err != nil {
				return err
			}
		}
		cmd := exec.Command(
			"sudo", "systemctl", "restart", "rollapp",
		)

		time.Sleep(time.Second * 2)
		WaitForHealthyRollApp("http://localhost:26657/health")
		_, err = bash.ExecCommandWithStdout(cmd)
		if err != nil {
			return err
		}
	} else {
		pterm.Info.Println("block time settings already up to date")
		pterm.Info.Println("restarting rollapp process to ensure correct block time is applied")
		err = servicemanager.RestartSystemdService("rollapp")
		if err != nil {
			return err
		}
		WaitForHealthyRollApp("http://localhost:26657/health")
	}

	return nil
}

func WaitForHealthyRollApp(url string) {
	timeout := time.After(20 * time.Second)
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	spinner, _ := pterm.DefaultSpinner.Start("waiting for rollapp to become healthy")

	for {
		select {
		case <-timeout:
			spinner.Fail("Timeout: Failed to receive expected response within 20 seconds")
			return
		case <-ticker.C:
			// nolint:gosec
			resp, err := http.Get(url)
			if err != nil {
				fmt.Printf("Error making request: %v\n", err)
				continue
			}

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				fmt.Printf("Error reading response body: %v\n", err)
				continue
			}
			// nolint:errcheck
			resp.Body.Close()

			var response RollappHealthResponse

			err = json.Unmarshal(body, &response)
			if err != nil {
				fmt.Printf("Error unmarshaling JSON: %v\n", err)
				continue
			}

			if response.Result.IsHealthy {
				spinner.Success("RollApp is healthy")
				return
			}
		}
	}
}
