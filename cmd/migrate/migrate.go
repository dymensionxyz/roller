package migrate

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
	"strings"
)

type VersionData struct {
	Major int
	Minor int
	Patch int
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
			versionData, err := GetVersionData(rlpCfg)
			if err != nil {
				utils.PrettifyErrorIfExists(err)
				return
			}

			if versionData.Major == 0 && versionData.Minor < 2 {
				switch {
				case versionData.Patch < 4:
					err := performMigration("v0.1.4", rlpCfg)
					utils.PrettifyErrorIfExists(err)
				}
			}

			fmt.Println("💈 Roller has migrated successfully to v0.1.4!")
		},
	}
	return cmd
}

func performMigration(version string, rlpCfg config.RollappConfig) error {
	switch version {
	case "v0.1.4":
		dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
		dymintCfg, err := toml.LoadFile(dymintTomlPath)
		if err != nil {
			return err
		}
		sequencer.EnableDymintMetrics(dymintCfg)
		return config.WriteTomlToFile(dymintTomlPath, dymintCfg)
	}
	return nil
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
