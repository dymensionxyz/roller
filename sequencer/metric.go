package sequencer

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/pelletier/go-toml"
	"os"
	"path/filepath"
)

func SetDefaultDymintConfig(root string, rlpCfg config.RollappConfig) error {
	dymintTomlPath := GetDymintFilePath(root)
	dymintCfg, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	damanager := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	daConfig := damanager.GetSequencerDAConfig()
	dymintCfg.Set("da_layer", rlpCfg.DA)
	if daConfig != "" {
		dymintCfg.Set("da_config", daConfig)
	}
	dymintCfg.Set("instrumentation.prometheus", true)
	dymintCfg.Set("instrumentation.prometheus_listen_addr", ":2112")
	hubKeysDir := filepath.Join(rlpCfg.Home, consts.ConfigDirName.HubKeys)
	dymintCfg.Set("settlement_layer", "dymension")
	dymintCfg.Set("block_batch_size", 500)
	dymintCfg.Set("namespace_id", "000000000000ffff")
	dymintCfg.Set("block_time", "0.2s")
	dymintCfg.Set("batch_submit_max_time", "100s")
	dymintCfg.Set("empty_blocks_max_time", "10s")
	dymintCfg.Set("settlement_config.rollapp_id", rlpCfg.RollappID)
	dymintCfg.Set("settlement_config.node_address", rlpCfg.HubData.RPC_URL)
	dymintCfg.Set("settlement_config.dym_account_name", consts.KeysIds.HubSequencer)
	dymintCfg.Set("settlement_config.keyring_home_dir", hubKeysDir)
	dymintCfg.Set("settlement_config.gas_prices", rlpCfg.HubData.GAS_PRICE+consts.Denoms.Hub)
	file, err := os.Create(dymintTomlPath)
	if err != nil {
		return err
	}
	_, err = file.WriteString(dymintCfg.String())
	if err != nil {
		return err
	}
	return nil
}
