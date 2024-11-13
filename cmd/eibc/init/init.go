package init

import (
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/templates"
)

//go:embed templates/*.tmpl
var embeddedTemplates embed.FS

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
			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)

			rollerConfig, err := roller.LoadConfig(rollerHome)
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

			isEibcClientInitialized, err := filesystem.DirNotEmpty(eibcHome)
			if err != nil {
				pterm.Error.Println("failed to check eibc client initialized", err)
				return
			}

			if isEibcClientInitialized {
				pterm.Warning.Println("eibc client already initialized")

				kc := eibcutils.GetKeyConfig()
				pterm.Info.Printfln("checking for existing %s address", kc.ID)
				ok, err := kc.IsInKeyring(home)
				if err != nil {
					pterm.Error.Println("failed to get eibc key info: ", err)
					return
				}

				if ok {
					pterm.Info.Println("eibc key already present in the keyring")
					pterm.Info.Println("checking for existing policies")
					pol, err := eibcutils.GetPolicies(eibcHome, hd)
					if err != nil {
						pterm.Error.Println("failed to get policies: ", err)
						return
					}

					if pol != nil && len(pol.GroupPolicies) > 0 {
						pterm.Info.Printfln("policies already present for %s", kc.ID)
						fmt.Println(pol.GroupPolicies[0].Address)
						return
					}
				}

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

			ki, err := eibcutils.EnsureWhaleAccount()
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
				rollerData, err = roller.LoadConfig(home)
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
				for {
					raID, _ = pterm.DefaultInteractiveTextInput.WithDefaultText("Please enter the RollApp ID to fulfill eibc orders for").
						Show()

					_, err := rollapp.ValidateChainID(raID)
					if err != nil {
						pterm.Error.Printf("'%s' is not a valid RollApp ID: %v\n", raID, err)
						continue
					} else {
						break
					}
				}
			}

			var fNodes []string
			var rpc string
			for {
				// Prompt the user for the RPC URL
				rpc, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"dymint rpc endpoint that you will provide (example: rpc.rollapp.dym.xyz)",
				).Show()
				if !strings.HasPrefix(rpc, "http://") && !strings.HasPrefix(rpc, "https://") {
					rpc = "https://" + rpc
				}

				isValid := config.IsValidURL(rpc)

				if !isValid {
					pterm.Error.Println("Invalid URL. Please try again.")
				} else {
					fNodes = append(fNodes, rpc)
					break
				}
			}

			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
			err = eibcutils.AddRollappToEibc(raID, eibcHome, fNodes)
			if err != nil {
				pterm.Error.Println("failed to add the rollapp to eibc config: ", err)
				return
			}

			err = updateEibcConfig(eibcConfigPath, hd)
			if err != nil {
				pterm.Error.Println("failed to update config", err)
				return
			}

			err = createGroupIfNotPresent(ki, hd, eibcHome)
			if err != nil {
				return
			}

			err = createPolicyIfNotPresent(hd, eibcHome)
			if err != nil {
				return
			}

			if err == nil {
				defer func() {
					pterm.Info.Println("next steps:")
					pterm.Info.Printf(
						"run %s to start the eibc client in interactive mode\n",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprintf("roller eibc start"),
					)
				}()
			}
		},
	}
	return cmd
}

func createGroupIfNotPresent(ki *keys.KeyInfo, hd consts.HubData, eibcHome string) error {
	grp, err := eibcutils.GetGroups(ki.Address, hd)
	if err != nil {
		pterm.Error.Println("failed to get groups: ", err)
		return nil
	}

	if grp == nil || len(grp.Groups) == 0 {
		err = createMembersFile(eibcHome, ki)
		if err != nil {
			pterm.Error.Println("failed to create members file: ", err)
			return nil
		}
		membersDefinitionFilePath := filepath.Join(eibcHome, "init", "members.json")
		cGrpCmd := eibcutils.GetCreateGroupPolicyCmd(
			eibcHome,
			"some",
			membersDefinitionFilePath,
			hd,
		)

		out, err := bash.ExecCommandWithStdout(cGrpCmd)
		if err != nil {
			pterm.Error.Println("failed to create group: ", err)
			return nil
		}

		fmt.Println(out.String())
	}
	return err
}

func createPolicyIfNotPresent(hd consts.HubData, eibcHome string) error {
	pol, err := eibcutils.GetPolicies(eibcHome, hd)
	if err != nil {
		return err
	}

	if pol == nil || len(pol.GroupPolicies) == 0 {
		err = createPolicyFile(eibcHome)
		if err != nil {
			pterm.Error.Println("failed to create members file: ", err)
			return nil
		}
		policyDefinitionFilePath := filepath.Join(eibcHome, "init", "policy.json")
		cPolicyCmd := eibcutils.GetCreateGroupPolicyCmd(
			eibcHome,
			"some",
			policyDefinitionFilePath,
			hd,
		)

		out, err := bash.ExecCommandWithStdout(cPolicyCmd)
		if err != nil {
			pterm.Error.Println("failed to create policy: ", err)
			return nil
		}

		fmt.Println(out.String())
		return err
	}

	j, err := json.MarshalIndent(pol, "", "  ")
	if err != nil {
		pterm.Error.Println("failed to marshal policy: ", err)
		return nil
	}

	fmt.Println(string(j))

	return err
}

func createMembersFile(eibcHome string, ki *keys.KeyInfo) error {
	tmplSrc := filepath.Join("templates", "members.json.tmpl")
	tmplDst := filepath.Join(eibcHome, "init", "members.json")

	err := templates.WriteToFile(
		tmplSrc,
		tmplDst,
		*ki,
		embeddedTemplates,
	)
	if err != nil {
		pterm.Error.Printfln(
			"failed to write %s template to file %s: %v",
			tmplSrc,
			tmplDst,
			err,
		)
		return err
	}
	return nil
}

func createPolicyFile(eibcHome string) error {
	tmplSrc := filepath.Join("templates", "policy.json.tmpl")
	tmplDst := filepath.Join(eibcHome, "init", "policy.json")

	t, err := embeddedTemplates.ReadFile(tmplSrc)
	if err != nil {
		pterm.Error.Printfln("failed to parse template: %v", err)
		return err
	}

	err = os.WriteFile(tmplDst, t, 0o644)
	if err != nil {
		pterm.Error.Printfln("failed to export template")
		return err
	}

	return nil
}

func updateEibcConfig(eibcConfigPath string, hd consts.HubData) error {
	updates := map[string]interface{}{
		"node_address":              hd.RpcUrl,
		"order_polling.indexer_url": consts.DefaultIndexer,
	}
	err := yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
	if err != nil {
		pterm.Error.Println("failed to update config", err)
		return err
	}

	return nil
}
