package migrate

import (
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
)

type VersionMigratorV0112 struct{}

func (v *VersionMigratorV0112) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 12
}

func (v *VersionMigratorV0112) PerformMigration(rlpCfg config.RollappConfig) error {
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	da := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	sequencerDaConfig := da.GetSequencerDAConfig()
	if sequencerDaConfig == "" {
		return nil
	}
	return utils.UpdateFieldInToml(dymintTomlPath, "da_config", sequencerDaConfig)
}
