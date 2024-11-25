package init

import (
	"embed"
	_ "embed"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/templates"
	"github.com/dymensionxyz/roller/utils/tx"
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
			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
			var ki *keys.KeyInfo

			rollerConfig, err := roller.LoadConfig(rollerHome)
			if err != nil || rollerConfig.HubData.ID == consts.MockHubID ||
				rollerConfig.HubData.ID == "" {
				pterm.Warning.Println("no roller config found")
				pterm.Info.Println("initializing for environment")

				envs := []string{"playground", "custom"}
				env, _ := pterm.DefaultInteractiveSelect.
					WithDefaultText(
						"select the environment you want to initialize eibc client for",
					).
					WithOptions(envs).
					Show()

				if env == "custom" {
					var rollerConfig roller.RollappConfig
					hdid, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide hub chain id").
						Show()
					hdrpc, _ := pterm.DefaultInteractiveTextInput.WithDefaultText("provide hub rpc endpoint").
						Show()

					rollerConfig.HubData.ID = hdid
					rollerConfig.HubData.RpcUrl = hdrpc

					hd = rollerConfig.HubData

					rollerCfgDir := roller.GetRootDir()
					err = os.MkdirAll(rollerCfgDir, 0o755)
					if err != nil {
						pterm.Error.Println("failed to create roller config dir", err)
						return
					}

					err := roller.WriteConfig(rollerConfig)
					if err != nil {
						pterm.Error.Println("failed to write roller config", err)
						return
					}
				} else {
					hd = consts.Hubs[env]
				}
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
					ki, err = kc.Info(home)
					if err != nil {
						pterm.Error.Println("failed to get eibc key info: ", err)
						return
					}

					pterm.Info.Println("eibc key already present in the keyring")
					pterm.Info.Println("checking for existing delegation groups")
					grp, err := eibcutils.GetGroups(eibcHome, ki.Address, hd)
					if err != nil {
						pterm.Error.Println("failed to get groups: ", err)
						return
					}

					if len(grp.Groups) > 0 {
						groupID := grp.Groups[0].ID
						pterm.Info.Printfln("delegation group found with ID: %s", groupID)

						pterm.Info.Println("checking for existing policies")
						pol, err := eibcutils.GetPolicies(eibcHome, groupID, hd)
						if err != nil {
							pterm.Error.Println("failed to get policies: ", err)
							return
						}

						if pol != nil && len(pol.GroupPolicies) > 0 {
							pterm.Info.Printfln("policies already present for %s", kc.ID)

							printPolicyAddress(pol.GroupPolicies[0].Address)

							updates := map[string]interface{}{
								"fulfillers.policy_address": pol.GroupPolicies[0].Address,
								"operator.group_id":         groupID,
							}
							err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
							if err != nil {
								pterm.Error.Println("failed to update config", err)
								return
							}

							return
						} else {
							policyAddr, err := createPolicyIfNotPresent(eibcHome, groupID, hd)
							if err != nil {
								pterm.Error.Println("failed to create policy: ", err)
								return
							}

							printPolicyAddress(policyAddr)
							updates := map[string]interface{}{
								"fulfillers.policy_address": policyAddr,
								"operator.group_id":         groupID,
							}
							err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
							if err != nil {
								pterm.Error.Println("failed to update config", err)
								return
							}
							return
						}
					}
				} else {
					pterm.Info.Printfln("safe to override %s", eibcHome)

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

					c := eibcutils.GetInitCmd()
					err = bash.ExecCmd(c)
					if err != nil {
						pterm.Error.Println("failed to initialize eibc client", err)
						return
					}

					ki, err = eibcutils.EnsureWhaleAccount()
					if err != nil {
						pterm.Error.Printf("failed to create whale account: %v\n", err)
						return
					}
				}
			} else {
				deps := dependencies.DefaultEibcClientPrebuiltDependencies()
				for _, v := range deps {
					err := dependencies.InstallBinaryFromRelease(v)
					if err != nil {
						pterm.Error.Printfln("failed to install binary: %s", err)
						return
					}
				}

				pterm.Info.Printfln("not initialized, initializing %s", eibcHome)
				c := eibcutils.GetInitCmd()
				err = bash.ExecCmd(c)
				if err != nil {
					pterm.Error.Println("failed to initialize eibc client", err)
					return
				}

				ki, err = eibcutils.EnsureWhaleAccount()
				if err != nil {
					pterm.Error.Printf("failed to create whale account: %v\n", err)
					return
				}
			}

			pterm.Info.Println(
				"you are about to run the eibc client for the following Dymension network:",
			)
			fmt.Println("network ID:",
				pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
					Sprint(hd.ID),
			)

			for {
				cqc := keys.ChainQueryConfig{
					Binary: consts.Executables.Dymension,
					Denom:  consts.Denoms.Hub,
					RPC:    hd.RpcUrl,
				}
				balance, err := keys.QueryBalance(cqc, ki.Address)
				if err != nil {
					pterm.Error.Println("failed to get balance: ", err)
					return
				}

				if !balance.Amount.IsPositive() {
					pterm.Info.Println(
						"please fund the addresses below to run the eibc client. this address will be the operator address of the client.",
					)
					ki.Print(keys.WithName(), keys.WithMnemonic())
					proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
						WithDefaultText(
							"press 'y' when the wallets are funded",
						).Show()
					if !proceed {
						pterm.Error.Println("cancelled by user")
						return
					}
				} else {
					break
				}
			}

			var runForExisting bool
			var raID string
			rollerConfigFilePath := filepath.Join(roller.GetRootDir(), consts.RollerConfigFileName)
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

				rollerData, err = roller.LoadConfig(roller.GetRootDir())
				if err != nil {
					pterm.Error.Printf("failed to load rollapp config: %v\n", err)
					return
				}
				rollerRaID := rollerData.RollappID
				rollerHubData := rollerData.HubData

				var rlyFromRoller bool
				if rollerRaID != "" {
					msg := fmt.Sprintf(
						"the retrieved RollApp ID is: %s, would you like to initialize the eibc client for this RollApp?",
						pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
							Sprint(rollerRaID),
					)
					rlyFromRoller, _ = pterm.DefaultInteractiveConfirm.WithDefaultText(msg).Show()
					if rlyFromRoller {
						raID = rollerRaID
						hd = rollerHubData
						runForExisting = true
					}
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
					"dymint rpc endpoint that you trust (example: rpc.rollapp.dym.xyz)",
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

			cfgBytes, err := os.ReadFile(eibcConfigPath)
			if err != nil {
				pterm.Error.Println("failed to read eibc config file: ", err)
				return
			}

			var cfg eibcutils.Config
			err = yaml.Unmarshal(cfgBytes, &cfg)
			if err != nil {
				pterm.Error.Println("failed to unmarshal eibc config file: ", err)
				return
			}

			cfg.RemoveChain("example_1234-1")
			updatedData, err := yaml.Marshal(&cfg)
			if err != nil {
				pterm.Error.Println("failed to marshal eibc config file: ", err)
				return
			}

			err = os.WriteFile(eibcConfigPath, updatedData, 0o644)
			if err != nil {
				pterm.Error.Println("failed to write eibc config file: ", err)
				return
			}

			gID, err := createGroupIfNotPresent(ki, hd, eibcHome)
			if err != nil {
				pterm.Error.Println("failed to create group: ", err)
				return
			}

			policyAddr, err := createPolicyIfNotPresent(eibcHome, gID, hd)
			if err != nil {
				pterm.Error.Println("failed to create policy: ", err)
				return
			}

			printPolicyAddress(policyAddr)
			updates := map[string]interface{}{
				"fulfillers.policy_address": policyAddr,
				"operator.group_id":         gID,
			}
			err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
			if err != nil {
				pterm.Error.Println("failed to update config", err)
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

func createGroupIfNotPresent(ki *keys.KeyInfo, hd consts.HubData, eibcHome string) (string, error) {
	grp, err := eibcutils.GetGroups(eibcHome, ki.Address, hd)
	if err != nil {
		pterm.Error.Println("failed to get groups: ", err)
		return "", err
	}

	if len(grp.Groups) > 0 {
		pterm.Info.Printfln("delegation group found with ID: %s", grp.Groups[0].ID)
		return grp.Groups[0].ID, nil
	}

	pterm.Info.Printfln("creating new delegation group for %s", ki.Address)
	var groupID string
	if grp == nil || len(grp.Groups) == 0 {
		err = createMembersFile(eibcHome, ki)
		if err != nil {
			pterm.Error.Println("failed to create members file: ", err)
			return "", err
		}
		membersDefinitionFilePath := filepath.Join(eibcHome, "init", "members.json")
		cGrpCmd := eibcutils.GetCreateGroupDelegationCmd(
			eibcHome,
			"eibc-operator",
			membersDefinitionFilePath,
			hd,
		)

		out, err := bash.ExecCommandWithStdout(cGrpCmd)
		if err != nil {
			pterm.Error.Println("failed to create group: ", err)
			return "", err
		}

		txHash, err := bash.ExtractTxHash(out.String())
		if err != nil {
			pterm.Error.Println("failed to extract tx hash: ", err)
			return "", err
		}

		err = tx.MonitorTransaction(hd.RpcUrl, txHash)
		if err != nil {
			return "", err
		}

		grp, err = eibcutils.GetGroups(eibcHome, ki.Address, hd)
		if err != nil {
			pterm.Error.Println("failed to get groups: ", err)
			return "", err
		}
		groupID = grp.Groups[0].ID
	}

	pterm.Info.Printfln("new delegation group created with ID: %s", groupID)
	return groupID, err
}

func createPolicyIfNotPresent(eibcHome, groupID string, hd consts.HubData) (string, error) {
	pol, err := eibcutils.GetPolicies(eibcHome, groupID, hd)
	if err != nil {
		return "", err
	}

	if len(pol.GroupPolicies) > 0 {
		pterm.Info.Printfln(
			"found existing policy for %s: %s",
			groupID,
			pol.GroupPolicies[0].Address,
		)
		return pol.GroupPolicies[0].Address, nil
	}

	if pol == nil || len(pol.GroupPolicies) == 0 {
		pterm.Info.Printfln("creating new delegation policy for group: %s", groupID)
		err = createPolicyFile(eibcHome)
		if err != nil {
			pterm.Error.Println("failed to create members file: ", err)
			return "", err
		}
		policyDefinitionFilePath := filepath.Join(eibcHome, "init", "policy.json")
		cPolicyCmd := eibcutils.GetCreateGroupPolicyCmd(
			eibcHome,
			"eibc-operator",
			policyDefinitionFilePath,
			groupID,
			hd,
		)

		out, err := bash.ExecCommandWithStdout(cPolicyCmd)
		if err != nil {
			pterm.Error.Println("failed to create policy: ", err)
			return "", err
		}

		txHash, err := bash.ExtractTxHash(out.String())
		if err != nil {
			pterm.Error.Println("failed to extract tx hash: ", err)
			return "", err
		}

		err = tx.MonitorTransaction(hd.RpcUrl, txHash)
		if err != nil {
			return "", err
		}

		pol, err := eibcutils.GetPolicies(eibcHome, groupID, hd)
		if err != nil {
			return "", err
		}

		return pol.GroupPolicies[0].Address, nil
	}

	return pol.GroupPolicies[0].Address, nil
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
		"order_polling.enabled":     true,
		"operator.account_name":     consts.KeysIds.Eibc,
		"gas.fees":                  "4000000000000000adym",
		// "rollapps.example_1234-1":   nil,
		"validation.interval":  "5m0s",
		"validation.wait_time": "61m0s",
	}
	err := yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
	if err != nil {
		pterm.Error.Println("failed to update config", err)
		return err
	}

	return nil
}

func printPolicyAddress(policyAddr string) {
	pterm.Info.Printfln("policy address: %s",
		pterm.DefaultBasicText.WithStyle(pterm.FgYellow.ToStyle()).
			Sprint(policyAddr),
	)
	pterm.Info.Println("share this with the LP provider")
}
