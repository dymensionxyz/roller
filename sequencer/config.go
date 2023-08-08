package sequencer

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils"
	"github.com/pelletier/go-toml"
	"os"
	"path/filepath"
	"strings"
)

func SetDefaultDymintConfig(rlpCfg config.RollappConfig) error {
	dymintTomlPath := GetDymintFilePath(rlpCfg.Home)
	dymintCfg, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	damanager := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	daConfig, err := damanager.GetSequencerDAConfig()
	if err != nil {
		return err
	}
	dymintCfg.Set("da_layer", string(rlpCfg.DA))
	if daConfig != "" {
		dymintCfg.Set("da_config", daConfig)
	}
	hubKeysDir := filepath.Join(rlpCfg.Home, consts.ConfigDirName.HubKeys)
	dymintCfg.Set("settlement_layer", "dymension")
	dymintCfg.Set("block_batch_size", "500")
	dymintCfg.Set("namespace_id", "000000000000ffff")
	dymintCfg.Set("block_time", "0.2s")
	dymintCfg.Set("batch_submit_max_time", "100s")
	dymintCfg.Set("empty_blocks_max_time", "10s")
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
	damanager := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	daConfig, err := damanager.GetSequencerDAConfig()
	if err != nil {
		return err
	}
	dymintCfg.Set("da_config", daConfig)
	return utils.WriteTomlTreeToFile(dymintCfg, dymintTomlPath)
}

func SetAppConfig(rlpCfg config.RollappConfig) error {
	appConfigFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	appCfg, err := toml.LoadFile(appConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", appConfigFilePath, err)
	}

	appCfg.Set("minimum-gas-prices", "0"+rlpCfg.Denom)
	appCfg.Set("api.enable", "true")

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

func GetRPCPort(rlpCfg config.RollappConfig) (string, error) {
	configFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "config.toml")
	addr, err := utils.GetKeyFromTomlFile(configFilePath, "rpc.laddr")
	if err != nil {
		return "", err
	}
	parts := strings.Split(addr, ":")
	port := parts[len(parts)-1]
	return port, nil
}

func GetRPCEndpoint(rlpCfg config.RollappConfig) (string, error) {
	rpcPort, err := GetRPCPort(rlpCfg)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("http://localhost:%s", rpcPort), nil
}
