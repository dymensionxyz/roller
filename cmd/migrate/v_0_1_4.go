package migrate

import (
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
)

type VersionMigratorV014 struct{}

func (v *VersionMigratorV014) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 4
}

func (v *VersionMigratorV014) PerformMigration(rlpCfg config.RollappConfig) error {
	return sequencer.SetDefaultDymintConfig(rlpCfg)
}
