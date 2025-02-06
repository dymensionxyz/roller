package create

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "create <height>",
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			height := args[0]
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			localRollerConfig, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			rollappConfig, err := rollapp.PopulateRollerConfigWithRaMetadataFromChain(
				home,
				localRollerConfig.RollappID,
				localRollerConfig.HubData,
			)
			errorhandling.PrettifyErrorIfExists(err)

			pterm.Info.Println(
				"in order to create a snapshot, please stop all the rollapp processes",
			)
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the rollapp process is stopped",
				).Show()
			if !proceed {
				pterm.Error.Println("cancelled by user")
				return
			}

			pterm.Info.Println(
				"ensuring the rollapp process is stopped",
			)

			servicesToStop := []string{"rollapp"}
			err = servicemanager.StopSystemServices(servicesToStop)
			if err != nil {
				pterm.Error.Println("failed to stop systemd services:", err)
				return
			}

			// approve the data directory deletion before take the snapshot
			rollappDirPath := filepath.Join(home, consts.ConfigDirName.Rollapp)

			dataDir := filepath.Join(rollappDirPath, "data")
			snapshotDir := filepath.Join(home, consts.ConfigDirName.Snapshots)

			if err := os.MkdirAll(snapshotDir, os.ModePerm); err != nil {
				pterm.Error.Println("failed to create folder:", err)
				return
			}

			privkeyFile := filepath.Join(rollappDirPath, "config", "priv_validator_key.json")
			backupFile := filepath.Join(snapshotDir, "priv_validator_key.json")

			err = os.Rename(privkeyFile, backupFile)
			if err != nil {
				pterm.Error.Println("failed to backup privkey:", err)
				return
			}

			if fi, err := os.Stat(dataDir); err == nil && fi.IsDir() {
				_, err := filesystem.DirNotEmpty(dataDir)
				if err != nil {
					pterm.Error.Printf("failed to check if data directory is empty: %v\n", err)
					os.Exit(1)
				}

				timestamp := time.Now().Format("2006-01-02-15-04-06")
				snapshotFileName := filepath.Join(snapshotDir, fmt.Sprintf("%s-%s-%s.tar.gz", rollappConfig.RollappID, height, timestamp))

				err = filesystem.CompressTarGz(dataDir, snapshotDir, snapshotFileName)
				if err != nil {
					pterm.Error.Println("failed to compress snapshot: ", err)
					return
				}
			}

			err = os.Rename(backupFile, privkeyFile)
			if err != nil {
				pterm.Error.Println("failed to restore privkey:", err)
				return
			}

			pterm.Info.Println(
				"data directory archive saved to ", snapshotDir,
			)
		},
	}
	return cmd
}
