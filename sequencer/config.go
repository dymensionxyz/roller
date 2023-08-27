package sequencer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/utils"
	"github.com/pelletier/go-toml"
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
	dymintCfg.Set("settlement_layer", "dymension")
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
	dymintCfg.Set("da_layer", string(rlpCfg.DA))
	daConfig := damanager.GetSequencerDAConfig()
	dymintCfg.Set("da_config", daConfig)
	if rlpCfg.DA == config.Celestia {
		celDAManager, ok := damanager.DataLayer.(*celestia.Celestia)
		if !ok {
			return fmt.Errorf("invalid damanager type, expected *celestia.Celestia, got %T", damanager.DataLayer)
		}
		dymintCfg.Set("namespace_id", celDAManager.NamespaceID)
	}
	return nil
}

func SetAppConfig(rlpCfg config.RollappConfig) error {
	appConfigFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	appCfg, err := toml.LoadFile(appConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", appConfigFilePath, err)
	}

	appCfg.Set("minimum-gas-prices", "0"+rlpCfg.Denom)
	appCfg.Set("api.enable", true)
	appCfg.Set("api.enabled-unsafe-cors", true)

	if appCfg.Has("json-rpc") {
		appCfg.Set("json-rpc.address", "0.0.0.0:8545")
		appCfg.Set("json-rpc.ws-address", "0.0.0.0:8546")
	}
	return utils.WriteTomlTreeToFile(appCfg, appConfigFilePath)
}

func SetTMConfig(rlpCfg config.RollappConfig) error {
	configFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "config.toml")
	var tomlCfg, err = toml.LoadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", configFilePath, err)
	}
	tomlCfg.Set("rpc.laddr", "tcp://0.0.0.0:26657")
	tomlCfg.Set("log_level", "debug")
	tomlCfg.Set("rpc.cors_allowed_origins", []string{"*"})
	return utils.WriteTomlTreeToFile(tomlCfg, configFilePath)
}

func (seq *Sequencer) ReadRPCPort() (string, error) {
	rpcAddr, err := seq.GetConfigValue("rpc.laddr")
	if err != nil {
		return "", err
	}
	parts := strings.Split(rpcAddr, ":")
	port := parts[len(parts)-1]
	return port, nil
}

func (seq *Sequencer) GetConfigValue(key string) (string, error) {
	configFilePath := filepath.Join(seq.RlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "config.toml")
	var tomlCfg, err = toml.LoadFile(configFilePath)
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
