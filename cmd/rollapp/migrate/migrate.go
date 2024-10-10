package migrate

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/migrations"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/upgrades"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Migrates the roller configuration to the newly installed version.",
		Run: func(cmd *cobra.Command, args []string) {
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerData, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			raUpgrade, err := upgrades.NewRollappUpgrade(string(rollerData.RollappVMType))
			if err != nil {
				pterm.Error.Println("failed to check rollapp version equality: ", err)
				return
			}
			pterm.Info.Printf(
				"starting migration process from %s to %s\n",
				rollerData.RollappBinaryVersion,
				raUpgrade.CurrentVersionCommit,
			)

			switch rollerData.RollappVMType {
			case consts.EVM_ROLLAPP:
				rollappType := "rollapp-evm"
				err = applyMigrations(
					rollerData.RollappBinaryVersion, raUpgrade.CurrentVersionCommit, rollappType,
					upgrades.EvmRollappUpgradeModules,
				)
				if err != nil {
					pterm.Error.Println("failed to apply migrations: ", err)
					return
				}
			case consts.WASM_ROLLAPP:
				rollappType := "rollapp-wasm"
				err = applyMigrations(
					rollerData.RollappBinaryVersion, raUpgrade.CurrentVersionCommit, rollappType,
					upgrades.WasmRollappUpgradeModules,
				)
				if err != nil {
					pterm.Error.Println("failed to apply migrations: ", err)
					return
				}
			default:
				pterm.Error.Println("unsupported rollapp VM type")
				return
			}
		},
	}
	return cmd
}

func applyMigrations(from, to, vmt string, versions []upgrades.Version) error {
	var fromTs time.Time
	var toTs time.Time
	var err error

	if strings.HasPrefix(from, "v") {
		fromTs, err = migrations.GetCommitTimestampByTag(
			"dymensionxyz",
			vmt,
			from,
		)
	} else {
		fromTs, err = migrations.GetCommitTimestamp(
			"dymensionxyz",
			vmt,
			from,
		)
	}
	if err != nil {
		return err
	}

	if strings.HasPrefix(from, "v") {
		toTs, err = migrations.GetCommitTimestampByTag(
			"dymensionxyz",
			vmt,
			to,
		)
	} else {
		toTs, err = migrations.GetCommitTimestamp(
			"dymensionxyz",
			vmt,
			to,
		)
	}
	if err != nil {
		return err
	}

	if fromTs == toTs || toTs.Before(fromTs) {
		return errors.New("the commit to migrate to must be younger then current commit")
	}

	var versionsToApply []upgrades.Version

	// get the commit timestamp of the version and compare it with the timestamp of the
	// from and to ts, if it's older than from ts, ignore it, if it's younger than to ts
	// ignore it, otherwise add it to the versions to apply
	pterm.Info.Println("iterating through all available upgrades")
	for _, v := range versions {
		var upgradeVersionTimestamp time.Time

		if strings.HasPrefix(from, "v") {
			upgradeVersionTimestamp, err = migrations.GetCommitTimestampByTag(
				"dymensionxyz",
				vmt,
				v.VersionIdentifier,
			)
			if err != nil {
				return err
			}
		} else {
			upgradeVersionTimestamp, err = migrations.GetCommitTimestamp(
				"dymensionxyz",
				vmt,
				v.VersionIdentifier,
			)
			if err != nil {
				return err
			}
		}

		fmt.Println("version: ", upgradeVersionTimestamp, v.VersionIdentifier)
		fmt.Println("from: ", fromTs, from)
		fmt.Println("to: ", fromTs, to)

		isNew := upgradeVersionTimestamp.Before(toTs) || upgradeVersionTimestamp.Equal(toTs)
		if upgradeVersionTimestamp.After(fromTs) || isNew {
			pterm.Info.Printf("adding %s to the relevant upgrades\n", v.VersionIdentifier)
			versionsToApply = append(versionsToApply, v)
		}
	}

	pterm.Info.Println("applying relevant upgrades")
	for _, version := range versionsToApply {
		pterm.Info.Printf("applying %s config changes\n", version.VersionIdentifier)
		err := applyMigration(version)
		if err != nil {
			pterm.Error.Printf(
				"failed to apply %s config changes: %v\n",
				version.VersionIdentifier,
				err,
			)
			return err
		}
	}
	return nil
}

func applyMigration(v upgrades.Version) error {
	// nested loops, yuck
	for _, module := range v.Modules {
		if len(module.Values.NewValues) != 0 {
			for _, nw := range module.Values.NewValues {
				err := tomlconfig.UpdateFieldInFile(module.ConfigFilePath, nw.Path, nw.Value)
				if err != nil {
					return err
				}
			}
		}

		if len(module.Values.DeprecatedValues) != 0 {
			for _, dw := range module.Values.DeprecatedValues {
				err := tomlconfig.RemoveFieldFromFile(module.ConfigFilePath, dw)
				if err != nil {
					return err
				}
			}
		}

		if len(module.Values.UpgradeableValues) != 0 {
			for _, uw := range module.Values.UpgradeableValues {
				err := tomlconfig.ReplaceFieldInFile(
					module.ConfigFilePath,
					uw.OldValuePath,
					uw.NewValuePath,
					uw.Value,
				)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}
