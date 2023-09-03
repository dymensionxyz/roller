package migrate

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
)

type VersionMigratorV0113 struct{}

func (v *VersionMigratorV0113) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 13
}

func (v *VersionMigratorV0113) PerformMigration(rlpCfg config.RollappConfig) error {
	// Update block time to 1 hour.
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	utils.UpdateFieldInToml(dymintTomlPath, "empty_blocks_max_time", "3600s")
	// Update relayer config path.
	if err := relayer.DeletePath(rlpCfg, "hub-rollapp"); err != nil {
		return err
	}
	return relayer.CreatePath(rlpCfg)
}
