package migrate

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/version"
	"github.com/spf13/cobra"
	"strings"
)

var migrationsRegistry = []VersionMigrator{
	&VersionMigratorV014{},
}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the roller configuration to the newly installed version.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			prevVersionData, err := GetPrevVersionData(rlpCfg)
			utils.PrettifyErrorIfExists(err)
			for _, migrator := range migrationsRegistry {
				if migrator.ShouldMigrate(*prevVersionData) {
					utils.PrettifyErrorIfExists(migrator.PerformMigration(rlpCfg))
				}
			}
			trimmedCurrentVersion := strings.Split(version.BuildVersion, "-")[0]
			fmt.Printf("💈 Roller has migrated successfully to %s!\n", trimmedCurrentVersion)
		},
	}
	return cmd
}

func GetPrevVersionData(rlpCfg config.RollappConfig) (*VersionData, error) {
	rollerPrevVersion := rlpCfg.RollerVersion
	var major, minor, patch int
	// Special case for the first version of roller, that didn't have a version field.
	if rollerPrevVersion == "" {
		return &VersionData{
			Major: 0,
			Minor: 1,
			Patch: 3,
		}, nil
	}
	trimmedVersionStr := strings.Split(rollerPrevVersion, "-")[0]
	n, err := fmt.Sscanf(trimmedVersionStr, "v%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return nil, err
	}
	if n != 3 {
		return nil, fmt.Errorf("failed to extract all version components from version %s",
			rollerPrevVersion)
	}
	return &VersionData{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

type VersionMigrator interface {
	PerformMigration(rlpCfg config.RollappConfig) error
	ShouldMigrate(prevVersion VersionData) bool
}

type VersionData struct {
	Major int
	Minor int
	Patch int
}
