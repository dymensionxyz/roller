package migrate

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
)

type VersionMigratorV0113 struct{}

func (v *VersionMigratorV0113) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 13
}

func (v *VersionMigratorV0113) PerformMigration(rlpCfg config.RollappConfig) error {
	if err := relayer.DeletePath(rlpCfg, "hub-rollapp"); err != nil {
		return err
	}
	return relayer.CreatePath(rlpCfg)
}
