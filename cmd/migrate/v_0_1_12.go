package migrate

import (
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/avail"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
	"path/filepath"
)

type VersionMigratorV0112 struct{}

func (v *VersionMigratorV0112) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 12
}

func (v *VersionMigratorV0112) PerformMigration(rlpCfg config.RollappConfig) error {
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	if rlpCfg.DA == "mock" {
		rlpCfg.DA = config.Local
		return config.WriteConfigToTOML(rlpCfg)
	}
	if rlpCfg.DA == config.Avail {
		availNewCfgPath := avail.GetCfgFilePath(rlpCfg.Home)
		if err := utils.MoveFile(filepath.Join(rlpCfg.Home, avail.ConfigFileName), availNewCfgPath); err != nil {
			return err
		}
	}
	da := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	sequencerDaConfig := da.GetSequencerDAConfig()
	if sequencerDaConfig == "" {
		return nil
	}
	if err := utils.UpdateFieldInToml(dymintTomlPath, "da_config", sequencerDaConfig); err != nil {
		return err
	}
	return nil
}
