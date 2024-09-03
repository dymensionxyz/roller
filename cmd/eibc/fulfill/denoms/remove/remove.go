package remove

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/eibc"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Commands to manage the whitelist of RollApps to fulfill eibc orders for",
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			home, err := os.UserHomeDir()
			if err != nil {
				pterm.Error.Println("failed to get user home dir", err)
				return
			}

			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
			isEibcClientInitialized, err := globalutils.DirNotEmpty(eibcHome)
			if err != nil {
				pterm.Error.Println("failed to check eibc client initialized", err)
				return
			}

			if !isEibcClientInitialized {
				pterm.Error.Println("eibc client not initialized")
				return
			}

			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
			data, err := os.ReadFile(eibcConfigPath)
			if err != nil {
				pterm.Error.Printf("Error reading file: %v\n", err)
				return
			}

			// Parse the YAML
			var config eibc.Config

			asset := args[0]
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				pterm.Error.Printf("Error reading file: %v\n", err)
				return
			}

			config.RemoveChain(asset)
			updatedData, err := yaml.Marshal(&config)
			fmt.Println(string(updatedData))
			if err != nil {
				pterm.Error.Printf("Error reading file: %v\n", err)
				return
			}

			err = os.WriteFile(eibcConfigPath, updatedData, 0o644)
			if err != nil {
				pterm.Error.Printf("Error reading file: %v\n", err)
				return
			}
		},
	}

	return cmd
}
