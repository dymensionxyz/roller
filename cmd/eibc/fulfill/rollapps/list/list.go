package list

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

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
			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
			isEibcClientInitialized, err := filesystem.DirNotEmpty(eibcHome)
			if err != nil {
				pterm.Error.Println("failed to check eibc client initialized", err)
				return
			}

			if !isEibcClientInitialized {
				pterm.Error.Println("eibc client not initialized")
				return
			}

			config, err := eibcutils.ReadConfig(eibcConfigPath)
			if err != nil {
				pterm.Error.Println("failed to read eibc config", err)
				return
			}

			for k, v := range config.Rollapps {
				fmt.Printf("%s requires %s validation(s):\n", k, v.MinConfirmations)
				for _, v := range v.FullNodes {
					fmt.Printf("\t%s\n", v)
				}
			}
		},
	}

	return cmd
}
