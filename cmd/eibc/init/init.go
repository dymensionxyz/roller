package init

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize eibc client",
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			rollerHome, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			home, err := os.UserHomeDir()
			if err != nil {
				pterm.Error.Println("failed to get user home dir", err)
				return
			}

			var hd consts.HubData
			rollerConfig, err := roller.LoadRollerConfig(rollerHome)
			if err != nil {
				pterm.Warning.Println("no roller config found")
				pterm.Info.Println("initializing for environment")

				envs := []string{"playground"}
				env, _ := pterm.DefaultInteractiveSelect.
					WithDefaultText(
						"select the environment you want to initialize eibc client for",
					).
					WithOptions(envs).
					Show()
				hd = consts.Hubs[env]
			} else {
				hd = rollerConfig.HubData
			}

			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
			isEibcClientInitialized, err := filesystem.DirNotEmpty(eibcHome)
			if err != nil {
				pterm.Error.Println("failed to check eibc client initialized", err)
				return
			}

			if isEibcClientInitialized {
				pterm.Warning.Println("eibc client already initialized")
				msg := fmt.Sprintf(
					"Directory %s is not empty. Do you want to overwrite it?",
					eibcHome,
				)
				shouldOverwrite, err := pterm.DefaultInteractiveConfirm.WithDefaultText(msg).
					WithDefaultValue(false).
					Show()
				if err != nil {
					errorhandling.PrettifyErrorIfExists(err)
					return
				}

				if shouldOverwrite {
					err = os.RemoveAll(eibcHome)
					if err != nil {
						errorhandling.PrettifyErrorIfExists(err)
						return
					}
					// nolint:gofumpt
					err = os.MkdirAll(eibcHome, 0o755)
					if err != nil {
						errorhandling.PrettifyErrorIfExists(err)
						return
					}
				} else {
					os.Exit(0)
				}
			}

			c := eibcutils.GetInitCmd()
			err = bash.ExecCmd(c)
			if err != nil {
				pterm.Error.Println("failed to initialize eibc client", err)
				return
			}

			err = eibcutils.EnsureWhaleAccount()
			if err != nil {
				pterm.Error.Printf("failed to create whale account: %v\n", err)
				return
			}

			var runForExisting bool
			var raID string
			rollerConfigFilePath := filepath.Join(home, consts.RollerConfigFileName)
			var rollerData roller.RollappConfig

			_, err = os.Stat(rollerConfigFilePath)
			if err != nil {
				if errors.Is(err, fs.ErrNotExist) {
					pterm.Info.Println("existing roller configuration not found")
					runForExisting = false
				} else {
					pterm.Error.Println("failed to check existing roller config")
					return
				}
			} else {
				pterm.Info.Println("existing roller configuration found, retrieving RollApp ID from it")
				rollerData, err = roller.LoadRollerConfig(home)
				if err != nil {
					pterm.Error.Printf("failed to load rollapp config: %v\n", err)
					return
				}
				rollerRaID := rollerData.RollappID
				rollerHubData := rollerData.HubData
				msg := fmt.Sprintf(
					"the retrieved RollApp ID is: %s, would you like to initialize the eibc client for this RollApp?",
					rollerRaID,
				)
				rlyFromRoller, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(msg).Show()
				if rlyFromRoller {
					raID = rollerRaID
					hd = rollerHubData
					runForExisting = true
				}

				if !rlyFromRoller {
					runForExisting = false
				}
			}

			if !runForExisting {
				raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText("Please enter the RollApp ID").
					Show()
			}

			raFeePercentage, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("Please provide the fee percentage for the RollApp").
				Show()

			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
			updates := map[string]interface{}{
				"node_address":              hd.RPC_URL,
				"whale.account_name":        consts.KeysIds.Eibc,
				"order_polling.interval":    "25s",
				"order_polling.indexer_url": "http://44.206.211.230:3000/",
				"order_polling.enabled":     true,
			}

			err = eibcutils.AddRollappToEibc(raFeePercentage, raID, eibcHome)
			if err != nil {
				pterm.Error.Println("failed to add the rollapp to eibc config: ", err)
				return
			}

			err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
			if err != nil {
				pterm.Error.Println("failed to update config", err)
				return
			}

			err = eibcutils.CreateMongoDbContainer()
			if err != nil {
				pterm.Error.Println("failed to create mongodb container:", err)
				return
			}
			pterm.Info.Println("created eibc mongodb container")

			pterm.Info.Println("eibc config updated successfully")
			pterm.Info.Printf(
				"eibc client initialized successfully at %s\n",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprintf(eibcHome),
			)

			defer func() {
				pterm.Info.Println("next steps:")
				pterm.Info.Printf(
					"run %s to start the eibc client in interactive mode\n",
					pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
						Sprintf("roller eibc start"),
				)
			}()
		},
	}
	return cmd
}
