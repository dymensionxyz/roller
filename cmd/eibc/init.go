package eibc

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v3"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

func initCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize eibc client",
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			rollerHome, err := globalutils.ExpandHomePath(
				cmd.Flag(utils.FlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			home, err := os.UserHomeDir()
			if err != nil {
				pterm.Error.Println("failed to get user home dir", err)
			}

			rollerConfig, err := tomlconfig.LoadRollerConfig(rollerHome)
			if err != nil {
				return
			}

			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
			isEibcClientInitialized, err := globalutils.DirNotEmpty(eibcHome)
			if err != nil {
				pterm.Error.Println("failed to check eibc client initialized", err)
				return
			}
			oh := initconfig.NewOutputHandler(false)

			if isEibcClientInitialized {
				pterm.Warning.Println("eibc client already initialized")
				shouldOverwrite, err := oh.PromptOverwriteConfig(eibcHome)
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

			c := GetInitCommand()

			err = bash.ExecCmd(c)
			if err != nil {
				return
			}

			err = ensureWhaleAccount()
			if err != nil {
				pterm.Error.Printf("failed to create whale account: %v\n", err)
				return
			}

			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
			node, err := yamlconfig.CreateYamlNodeFromFile(eibcHome, "config.yaml")
			if err != nil {
				fmt.Printf("failed to retrieve yaml config: %v\n", err)
				return
			}

			// Get the actual content node (usually the first child of the document node)
			contentNode := &node
			if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
				contentNode = &node.Content[0]
			}

			// Update the nested fields
			err = yamlconfig.UpdateNestedYAML(
				*contentNode,
				[]string{"node_address"},
				rollerConfig.HubData.RPC_URL,
			)
			if err != nil {
				fmt.Printf("Error updating YAML: %v\n", err)
				return
			}

			err = yamlconfig.UpdateNestedYAML(
				*contentNode,
				[]string{"whale", "account_name"},
				consts.KeysIds.Eibc,
			)
			if err != nil {
				fmt.Printf("Error updating YAML: %v\n", err)
				return
			}

			// Marshal the updated YAML
			updatedData, err := yaml.Marshal(&node)
			if err != nil {
				fmt.Printf("Error marshaling YAML: %v\n", err)
				return
			}

			// Write the updated YAML back to the original file
			err = os.WriteFile(eibcConfigPath, updatedData, 0o644)
			if err != nil {
				fmt.Printf("Error writing file: %v\n", err)
				return
			}

			pterm.Info.Println("eibc config updated successfully")
		},
	}
	return cmd
}

func GetInitCommand() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Eibc,
		"init",
	)
	return cmd
}
