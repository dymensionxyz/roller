package sequencer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"

	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config"
)

func SetDefaultDymintConfig(rlpCfg config.RollappConfig) error {
	dymintTomlPath := GetDymintFilePath(rlpCfg.Home)
	dymintCfg, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	if err := updateDaConfigInToml(rlpCfg, dymintCfg); err != nil {
		return err
	}

	hubKeysDir := filepath.Join(rlpCfg.Home, consts.ConfigDirName.HubKeys)
	dymintCfg.Set("max_idle_time", "1h0m0s")
	dymintCfg.Set("max_proof_time", "10s")

	if rlpCfg.HubData.ID == consts.MockHubID {
		dymintCfg.Set("settlement_layer", "mock")
	} else {
		dymintCfg.Set("settlement_layer", "dymension")
	}

	dymintCfg.Set("block_batch_size", "500")
	dymintCfg.Set("block_time", "0.2s")
	dymintCfg.Set("batch_submit_max_time", "100s")
	dymintCfg.Set("empty_blocks_max_time", "3600s")
	dymintCfg.Set("rollapp_id", rlpCfg.RollappID)
	dymintCfg.Set("node_address", rlpCfg.HubData.RPC_URL)
	dymintCfg.Set("dym_account_name", consts.KeysIds.HubSequencer)
	dymintCfg.Set("keyring_home_dir", hubKeysDir)
	dymintCfg.Set("gas_prices", rlpCfg.HubData.GAS_PRICE+consts.Denoms.Hub)
	dymintCfg.Set("instrumentation.prometheus", true)
	dymintCfg.Set("instrumentation.prometheus_listen_addr", ":2112")
	dymintCfg.Set("batch_submit_max_time", "1h0m0s")

	file, err := os.Create(dymintTomlPath)
	if err != nil {
		return err
	}
	_, err = file.WriteString(dymintCfg.String())
	return err
}

func UpdateDymintDAConfig(rlpCfg config.RollappConfig) error {
	dymintTomlPath := GetDymintFilePath(rlpCfg.Home)
	dymintCfg, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	if err := updateDaConfigInToml(rlpCfg, dymintCfg); err != nil {
		return err
	}
	return utils.WriteTomlTreeToFile(dymintCfg, dymintTomlPath)
}

func updateDaConfigInToml(rlpCfg config.RollappConfig, dymintCfg *toml.Tree) error {
	damanager := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	dymintCfg.Set("da_layer", "mock")
	// daConfig := damanager.GetSequencerDAConfig()
	// dymintCfg.Set("da_config", daConfig)
	if rlpCfg.DA == consts.Celestia {
		celDAManager, ok := damanager.DataLayer.(*celestia.Celestia)
		if !ok {
			return fmt.Errorf(
				"invalid damanager type, expected *celestia.Celestia, got %T",
				damanager.DataLayer,
			)
		}
		dymintCfg.Set("namespace_id", celDAManager.NamespaceID)
	}

	if rlpCfg.DA == consts.Local {
		dymintCfg.Set("da_layer", "mock")
	}

	return nil
}

func SetAppConfig(rlpCfg config.RollappConfig) error {
	appConfigFilePath := filepath.Join(getSequencerConfigDir(rlpCfg.Home), "app.toml")
	appCfg, err := toml.LoadFile(appConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", appConfigFilePath, err)
	}

	appCfg.Set("minimum-gas-prices", "1000000000"+rlpCfg.Denom)
	appCfg.Set("api.enable", true)
	appCfg.Set("api.enabled-unsafe-cors", true)

	if appCfg.Has("json-rpc") {
		appCfg.Set("json-rpc.address", "0.0.0.0:8545")
		appCfg.Set("json-rpc.ws-address", "0.0.0.0:8546")
	}
	return utils.WriteTomlTreeToFile(appCfg, appConfigFilePath)
}

func SetTMConfig(rlpCfg config.RollappConfig) error {
	configFilePath := filepath.Join(getSequencerConfigDir(rlpCfg.Home), "config.toml")
	tomlCfg, err := toml.LoadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", configFilePath, err)
	}
	tomlCfg.Set("rpc.laddr", "tcp://0.0.0.0:26657")
	tomlCfg.Set("rpc.timeout_broadcast_tx_commit", "30s")
	tomlCfg.Set("rpc.max_subscriptions_per_client", "10")
	tomlCfg.Set("log_level", "debug")
	tomlCfg.Set("rpc.cors_allowed_origins", []string{"*"})
	return utils.WriteTomlTreeToFile(tomlCfg, configFilePath)
}

func (seq *Sequencer) ReadPorts() error {
	rpcAddr, err := seq.GetConfigValue("rpc.laddr")
	if err != nil {
		return err
	}

	seq.RPCPort = getPortFromAddress(rpcAddr)
	appCfg, err := toml.LoadFile(filepath.Join(getSequencerConfigDir(seq.RlpCfg.Home), "app.toml"))
	if err != nil {
		return err
	}

	jsonRpcAddr := appCfg.Get("json-rpc.address")
	seq.JsonRPCPort = getPortFromAddress(fmt.Sprint(jsonRpcAddr))
	apiAddr := appCfg.Get("api.address")
	seq.APIPort = getPortFromAddress(fmt.Sprint(apiAddr))
	return nil
}

func (seq *Sequencer) GetConfigValue(key string) (string, error) {
	configFilePath := filepath.Join(getSequencerConfigDir(seq.RlpCfg.Home), "config.toml")
	tomlCfg, err := toml.LoadFile(configFilePath)
	if err != nil {
		return "", fmt.Errorf("failed to load %s: %v", configFilePath, err)
	}
	value := tomlCfg.Get(key)
	if value == nil {
		return "", fmt.Errorf("failed to get value for key: %s", key)
	}
	return fmt.Sprint(value), nil
}

func (seq *Sequencer) GetRPCEndpoint() string {
	return "http://localhost:" + seq.RPCPort
}

func (seq *Sequencer) GetLocalEndpoint(port string) string {
	return "http://localhost:" + port
}

func getSequencerConfigDir(rollerHome string) string {
	return filepath.Join(rollerHome, consts.ConfigDirName.Rollapp, "config")
}

func getPortFromAddress(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[len(parts)-1]
}
