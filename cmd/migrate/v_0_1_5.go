package migrate

import (
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config"
)

type VersionMigratorV015 struct{}

func (v *VersionMigratorV015) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 5
}

func (v *VersionMigratorV015) PerformMigration(rlpCfg config.RollappConfig) error {
	if err := sequencer.SetAppConfig(rlpCfg); err != nil {
		return err
	}
	return sequencer.SetTMConfig(rlpCfg)
}
