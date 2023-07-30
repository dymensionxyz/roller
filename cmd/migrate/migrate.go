package migrate

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/version"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"strings"
)

type VersionMigrator interface {
	PerformMigration(rlpCfg config.RollappConfig) error
}

func NewVersionMigrator(version string) VersionMigrator {
	switch version {
	case "v0.1.4":
		return &VersionMigratorV014{}
	default:
		return nil
	}
}

type VersionData struct {
	Major int
	Minor int
	Patch int
}

type VersionMigratorV014 struct{}

func (v *VersionMigratorV014) PerformMigration(rlpCfg config.RollappConfig) error {
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	dymintCfg, err := toml.LoadFile(dymintTomlPath)
	if err != nil {
		return err
	}
	sequencer.EnableDymintMetrics(dymintCfg)
	return config.WriteTomlToFile(dymintTomlPath, dymintCfg)
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the roller configuration to the newly installed version.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			if err != nil {
				utils.PrettifyErrorIfExists(err)
				return
			}
			prevVersionData, err := GetVersionData(rlpCfg)
			if err != nil {
				utils.PrettifyErrorIfExists(err)
				return
			}
			if prevVersionData.Major < 1 {
				if prevVersionData.Minor < 2 {
					if prevVersionData.Patch < 4 {
						migrator := NewVersionMigrator("v0.1.4")
						if migrator != nil {
							err := migrator.PerformMigration(rlpCfg)
							utils.PrettifyErrorIfExists(err)
						}
					}
				}
			}
			trimmedCurrentVersion := strings.Split(version.BuildVersion, "-")[0]
			fmt.Printf("💈 Roller has migrated successfully to %s!\n", trimmedCurrentVersion)
		},
	}
	return cmd
}

func GetVersionData(rlpCfg config.RollappConfig) (*VersionData, error) {
	rollerPrevVersion := rlpCfg.RollerVersion
	var major, minor, patch int
	if rollerPrevVersion == "" {
		major, minor, patch = 0, 1, 3
	} else {
		trimmedVersionStr := strings.Split(rollerPrevVersion, "-")[0]
		n, err := fmt.Sscanf(trimmedVersionStr, "v%d.%d.%d", &major, &minor, &patch)
		if err != nil {
			return nil, err
		}
		if n != 3 {
			return nil, fmt.Errorf("failed to extract all version components from version %s",
				rollerPrevVersion)
		}
	}
	return &VersionData{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}
