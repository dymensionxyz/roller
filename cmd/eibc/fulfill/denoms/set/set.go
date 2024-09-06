package set

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <ibc-denom-id> <value>",
		Short: "Commands to manage the whitelist of RollApps to fulfill eibc orders for",
		Args:  cobra.ExactArgs(2),
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
			var node yaml.Node
			ibcDenom := args[0]
			value := args[1]

			if !strings.HasPrefix(ibcDenom, "ibc/") {
				pterm.Error.Println("invalid ibc denom")
				return
			}

			valueFloat, err := strconv.ParseFloat(value, 32)
			if err != nil {
				pterm.Error.Println("failed to convert value to float", err)
				return
			}
			err = yaml.Unmarshal(data, &node)
			if err != nil {
				pterm.Error.Println("failed to unmarshal config.yaml")
				return
			}

			// Get the actual content node (usually the first child of the document node)
			updates := map[string]interface{}{
				fmt.Sprintf("fulfill_criteria.min_fee_percentage.asset.%s", ibcDenom): valueFloat,
			}
			err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
			if err != nil {
				pterm.Error.Println("failed to update config", err)
			}
		},
	}

	return cmd
}
