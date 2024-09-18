package remove

import (
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <rollapp-id>",
		Short: "Commands to manage the whitelist of RollApps to fulfill eibc orders for",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			home, err := os.UserHomeDir()
			if err != nil {
				pterm.Error.Println("failed to get user home dir", err)
				return
			}

			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
			isEibcClientInitialized, err := filesystem.DirNotEmpty(eibcHome)
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

			rollAppID := args[0]
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				pterm.Error.Printf("Error reading file: %v\n", err)
				return
			}

			config.RemoveChain(rollAppID)
			updatedData, err := yaml.Marshal(&config)
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
