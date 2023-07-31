package migrate

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
)

type VersionMigratorV014 struct{}

func (v *VersionMigratorV014) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 4
}

func (v *VersionMigratorV014) PerformMigration(rlpCfg config.RollappConfig) error {
	return sequencer.SetDefaultDymintConfig(rlpCfg)
}
