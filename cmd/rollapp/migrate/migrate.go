package migrate

import (
	"fmt"

	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/version"
)

var migrationsRegistry = []VersionMigrator{}

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the roller configuration to the newly installed version.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rlpCfg, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			prevVersionData, err := GetPrevVersionData(rlpCfg)
			errorhandling.PrettifyErrorIfExists(err)
			for _, migrator := range migrationsRegistry {
				if migrator.ShouldMigrate(*prevVersionData) {
					errorhandling.PrettifyErrorIfExists(migrator.PerformMigration(rlpCfg))
				}
			}
			trimmedCurrentVersion := version.TrimVersionStr(version.BuildVersion)
			rlpCfg.RollerVersion = trimmedCurrentVersion
			err = roller.WriteConfig(rlpCfg)
			errorhandling.PrettifyErrorIfExists(err)
			fmt.Printf("💈 Roller has migrated successfully to %s!\n", trimmedCurrentVersion)
		},
	}
	return cmd
}

func GetPrevVersionData(rlpCfg roller.RollappConfig) (*VersionData, error) {
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

	trimmedVersionStr := version.TrimVersionStr(rollerPrevVersion)
	n, err := fmt.Sscanf(trimmedVersionStr, "v%d.%d.%d", &major, &minor, &patch)
	if err != nil {
		return nil, err
	}
	if n != 3 {
		return nil, fmt.Errorf(
			"failed to extract all version components from version %s",
			rollerPrevVersion,
		)
	}
	return &VersionData{
		Major: major,
		Minor: minor,
		Patch: patch,
	}, nil
}

type VersionMigrator interface {
	PerformMigration(rlpCfg roller.RollappConfig) error
	ShouldMigrate(prevVersion VersionData) bool
}

type VersionData struct {
	Major int
	Minor int
	Patch int
}
