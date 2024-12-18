package sequencer

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

func SetDefaultDymintConfig(rlpCfg roller.RollappConfig) error {
	dymintTomlPath := sequencerutils.GetDymintFilePath(rlpCfg.Home)
	dymintCfg, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	if err := updateDaConfigInToml(rlpCfg, dymintCfg); err != nil {
		return err
	}

	hubKeysDir := filepath.Join(rlpCfg.Home, consts.ConfigDirName.HubKeys)
	dymintCfg.Set("max_idle_time", "1h0m0s")
	dymintCfg.Set("max_proof_time", "5s")

	if rlpCfg.HubData.ID == consts.MockHubID {
		dymintCfg.Set("settlement_layer", "mock")
	} else {
		dymintCfg.Set("settlement_layer", "dymension")
	}

	dymintCfg.Set("block_batch_size", "500")
	dymintCfg.Set("block_time", "0.2s")
	dymintCfg.Set("settlement_node_address", rlpCfg.HubData.RpcUrl)
	dymintCfg.Set("dym_account_name", consts.KeysIds.HubSequencer)
	dymintCfg.Set("keyring_home_dir", hubKeysDir)
	dymintCfg.Set("keyring_backend", string(rlpCfg.KeyringBackend))
	dymintCfg.Set("gas_prices", rlpCfg.HubData.GasPrice+consts.Denoms.Hub)
	dymintCfg.Set("instrumentation.prometheus", true)
	dymintCfg.Set("instrumentation.prometheus_listen_addr", ":2112")
	dymintCfg.Set("batch_submit_time", "1h0m0s")

	file, err := os.Create(dymintTomlPath)
	if err != nil {
		return err
	}
	_, err = file.WriteString(dymintCfg.String())
	return err
}

func UpdateDymintDAConfig(rlpCfg roller.RollappConfig) error {
	dymintTomlPath := sequencerutils.GetDymintFilePath(rlpCfg.Home)
	dymintCfg, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	if err := updateDaConfigInToml(rlpCfg, dymintCfg); err != nil {
		return err
	}
	return tomlconfig.WriteTomlTreeToFile(dymintCfg, dymintTomlPath)
}

func updateDaConfigInToml(rlpCfg roller.RollappConfig, dymintCfg *toml.Tree) error {
	damanager := datalayer.NewDAManager(rlpCfg.DA.Backend, rlpCfg.Home, rlpCfg.KeyringBackend)
	dymintCfg.Set("da_layer", "mock")
	// daConfig := damanager.GetSequencerDAConfig()
	// dymintCfg.Set("da_config", daConfig)
	if rlpCfg.DA.Backend == consts.Celestia {
		celDAManager, ok := damanager.DataLayer.(*celestia.Celestia)
		if !ok {
			return fmt.Errorf(
				"invalid damanager type, expected *celestia.Celestia, got %T",
				damanager.DataLayer,
			)
		}
		dymintCfg.Set("namespace_id", celDAManager.NamespaceID)
	}

	if rlpCfg.DA.Backend == consts.Local {
		dymintCfg.Set("da_layer", "mock")
	}

	return nil
}

func SetAppConfig(rlpCfg roller.RollappConfig) error {
	appConfigFilePath := filepath.Join(
		sequencerutils.GetSequencerConfigDir(rlpCfg.Home),
		"app.toml",
	)
	appCfg, err := toml.LoadFile(appConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", appConfigFilePath, err)
	}

	as, err := genesis.GetAppStateFromGenesisFile(rlpCfg.Home)
	if err != nil {
		return err
	}

	var minimumGasPrice string
	if as.FeeMarket != nil && as.FeeMarket.Params != nil && as.FeeMarket.Params.MinGasPrice != "" {
		pterm.Info.Println("applying feemarket gas price")
		minimumGasPrice = as.FeeMarket.Params.MinGasPrice
	} else if len(as.RollappParams.Params.MinGasPrices) > 0 {
		pterm.Info.Println("applying rollappparam gas price")
		minimumGasPrice = as.RollappParams.Params.MinGasPrices[0].Amount.String()
	} else {
		pterm.Info.Println("applying default gas price")
		minimumGasPrice = consts.DefaultMinGasPrice
	}

	appCfg.Set("minimum-gas-prices", fmt.Sprintf("%s%s", minimumGasPrice, rlpCfg.BaseDenom))
	appCfg.Set("gas-adjustment", 1.3)
	appCfg.Set("api.enable", true)
	appCfg.Set("api.enabled-unsafe-cors", true)

	if appCfg.Has("json-rpc") {
		appCfg.Set("json-rpc.address", "0.0.0.0:8545")
		appCfg.Set("json-rpc.ws-address", "0.0.0.0:8546")
	}
	return tomlconfig.WriteTomlTreeToFile(appCfg, appConfigFilePath)
}

func SetTMConfig(rlpCfg roller.RollappConfig) error {
	configFilePath := filepath.Join(
		sequencerutils.GetSequencerConfigDir(rlpCfg.Home),
		"config.toml",
	)
	tomlCfg, err := toml.LoadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", configFilePath, err)
	}
	tomlCfg.Set("rpc.laddr", "tcp://0.0.0.0:26657")
	tomlCfg.Set("rpc.timeout_broadcast_tx_commit", "30s")
	tomlCfg.Set("rpc.max_subscriptions_per_client", "10")
	tomlCfg.Set("log_level", "debug")
	tomlCfg.Set("rpc.cors_allowed_origins", []string{"*"})
	return tomlconfig.WriteTomlTreeToFile(tomlCfg, configFilePath)
}

func (seq *Sequencer) ReadPorts() error {
	rpcAddr, err := seq.GetConfigValue("rpc.laddr")
	if err != nil {
		return err
	}

	seq.RPCPort = getPortFromAddress(rpcAddr)
	appCfg, err := toml.LoadFile(
		filepath.Join(sequencerutils.GetSequencerConfigDir(seq.RlpCfg.Home), "app.toml"),
	)
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
	configFilePath := filepath.Join(
		sequencerutils.GetSequencerConfigDir(seq.RlpCfg.Home),
		"config.toml",
	)
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

func getPortFromAddress(addr string) string {
	parts := strings.Split(addr, ":")
	return parts[len(parts)-1]
}
