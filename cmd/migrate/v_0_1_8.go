package migrate

import (
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config"
)

type VersionMigratorV018 struct{}

func (v *VersionMigratorV018) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 8
}

func (v *VersionMigratorV018) PerformMigration(rlpCfg config.RollappConfig) error {
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	return utils.UpdateFieldInToml(dymintTomlPath, "empty_blocks_max_time", "60s")
}
