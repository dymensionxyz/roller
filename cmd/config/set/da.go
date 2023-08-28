package set

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	global_utils "github.com/dymensionxyz/roller/utils"
	"os"
	"path/filepath"
)

func setDA(rlpCfg config.RollappConfig, value string) error {
	daValue := config.DAType(value)
	if daValue == rlpCfg.DA {
		return nil
	}
	supportedDas := []config.DAType{config.Celestia, config.Avail, config.Local}
	if !config.IsValidDAType(value) {
		return fmt.Errorf("invalid DA type. Supported types are: %v", supportedDas)
	}
	return updateDaConfig(rlpCfg, daValue)
}

func updateDaConfig(rlpCfg config.RollappConfig, newDa config.DAType) error {
	daCfgDirPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.DALightNode)
	dirExist, err := global_utils.DirNotEmpty(daCfgDirPath)
	if err != nil {
		return err
	}
	if dirExist {
		if yes, err := global_utils.PromptBool("Changing DA will remove the old DA keys permanently. Are you sure you want to proceed"); err != nil {
			return err
		} else if !yes {
			return nil
		}
	}
	if err := os.RemoveAll(daCfgDirPath); err != nil {
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
