package migrate

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/version"
)

type VersionMigratorV015 struct{}

func (v *VersionMigratorV015) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 5
}

func (v *VersionMigratorV015) PerformMigration(rlpCfg config.RollappConfig) error {
	err := sequencer.SetDefaultDymintConfig(rlpCfg)
	if err != nil {
		return err
	}
	rlpCfg.RollerVersion = trimVersionStr(version.BuildVersion)
	return config.WriteConfigToTOML(rlpCfg)
}
