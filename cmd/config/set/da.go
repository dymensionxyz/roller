package set

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	"os"
	"path/filepath"
)

func setDA(rlpCfg config.RollappConfig, value config.DAType) error {
	if value == rlpCfg.DA {
		return nil
	}
	supportedDas := []config.DAType{config.Celestia, config.Avail, config.Mock}
	for _, da := range supportedDas {
		if da == value {
			return updateDaConfig(rlpCfg, value)
		}
	}
	return fmt.Errorf("invalid DA type. Supported types are: %v", supportedDas)
}

func updateDaConfig(rlpCfg config.RollappConfig, newDa config.DAType) error {
	if err := cleanDADir(rlpCfg); err != nil {
		return err
	}
	daManager := datalayer.NewDAManager(newDa, rlpCfg.Home)
	if err := daManager.InitializeLightNodeConfig(); err != nil {
		return err
	}
	rlpCfg.DA = newDa
	if err := sequencer.UpdateDymintDAConfig(rlpCfg); err != nil {
		return err
	}
	return config.WriteConfigToTOML(rlpCfg)
}

func cleanDADir(cfg config.RollappConfig) error {
	return os.RemoveAll(filepath.Join(cfg.Home, consts.ConfigDirName.DALightNode))
}
