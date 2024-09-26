package list

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "Commands to manage the whitelist of RollApps to fulfill eibc orders for",
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
			var config eibcutils.Config
			err = yaml.Unmarshal(data, &config)
			if err != nil {
				pterm.Error.Printf("Error reading file: %v\n", err)
				return
			}

			for chain, percentage := range config.FulfillCriteria.MinFeePercentage.Chain {
				fmt.Printf("%s: %.6f\n", chain, percentage)
			}
		},
	}

	return cmd
}
