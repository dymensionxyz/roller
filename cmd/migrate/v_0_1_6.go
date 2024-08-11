package migrate

import (
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
)

type VersionMigratorV016 struct{}

func (v *VersionMigratorV016) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 6
}

func (v *VersionMigratorV016) PerformMigration(rlpCfg config.RollappConfig) error {
	return sequencer.SetTMConfig(rlpCfg)
}
