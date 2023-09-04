package migrate

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
)

type VersionMigratorV0113 struct{}

func (v *VersionMigratorV0113) ShouldMigrate(prevVersion VersionData) bool {
	return true
	//return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 13
}

func (v *VersionMigratorV0113) PerformMigration(rlpCfg config.RollappConfig) error {
	// Update block time to 1 hour.
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	if err := utils.UpdateFieldInToml(dymintTomlPath, "empty_blocks_max_time", "3600s"); err != nil {
		return err
	}
	return revertRlyPath(rlpCfg)
}

func revertRlyPath(rlpCfg config.RollappConfig) error {
	rlyCfg, err := relayer.ReadRlyConfig(rlpCfg)
	if err != nil {
		return err
	}
	srcData, err := utils.GetNestedValue(rlyCfg, []string{"paths", "hub-rollapp", "src"})
	if err != nil {
		return err
	}
	dstData, err := utils.GetNestedValue(rlyCfg, []string{"paths", "hub-rollapp", "dst"})
	if err != nil {
		return err
	}
	pathCfg, err := utils.GetNestedValue(rlyCfg, []string{"paths", "hub-rollapp"})
	if err != nil {
		return err
	}
	pathCfgMap := pathCfg.(map[interface{}]interface{})
	if err := utils.SetNestedValue(pathCfgMap, []string{"src"}, dstData); err != nil {
		return err
	}
	if err := utils.SetNestedValue(pathCfgMap, []string{"dst"}, srcData); err != nil {
		return err
	}
	if err := utils.SetNestedValue(rlyCfg, []string{"paths", "rollapp-hub"}, pathCfgMap); err != nil {
		return err
	}
	if err := utils.SetNestedValue(rlyCfg, []string{"paths", "hub-rollapp"}, nil); err != nil {
		return err
	}
	return relayer.WriteRlyConfig(rlpCfg, rlyCfg)
}
