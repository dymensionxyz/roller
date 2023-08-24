package migrate

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
)

type VersionMigratorV0111 struct{}

func (v *VersionMigratorV0111) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 9
}

func (v *VersionMigratorV0111) PerformMigration(rlpCfg config.RollappConfig) error {
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	return utils.UpdateFieldInToml(dymintTomlPath, "empty_blocks_max_time", "3600s")
}
