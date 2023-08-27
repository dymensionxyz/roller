package set

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"os"
	"path/filepath"
)

func setDA(rlpCfg config.RollappConfig, value config.DAType) error {
	if value == rlpCfg.DA {
		return nil
	}
	switch value {
	case config.Celestia:
		if err := cleanDADir(rlpCfg); err != nil {
			return err
		}
		daManager := datalayer.NewDAManager(value, rlpCfg.Home)
		if err := daManager.InitializeLightNodeConfig(); err != nil {
			return err
		}
		rlpCfg.DA = value
		return config.WriteConfigToTOML(rlpCfg)
	case config.Avail:
		return nil
	case config.Mock:
		return nil
	default:
		return fmt.Errorf("invalid DA type. Supported types are: %v", []string{string(config.Celestia),
			string(config.Avail), string(config.Mock)})
	}
}

func cleanDADir(cfg config.RollappConfig) error {
	return os.RemoveAll(filepath.Join(cfg.Home, consts.ConfigDirName.DALightNode))
}
