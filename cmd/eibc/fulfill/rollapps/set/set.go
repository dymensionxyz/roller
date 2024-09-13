package set

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <rollapp-id> <fee-percentage>",
		Short: "Commands to manage the whitelist of RollApps to fulfill eibc orders for",
		Long: `Commands to manage the whitelist of RollApps to fulfill eibc orders for

The fee-percentage is a float number between 0 and 100 which represents
the minimal percentage of the order fee that you want to receive for fulfilling an order.
Assume there's an eibc order for 100<token> with a fee of 3<token>,
if the percentage is set to 4, this order will be ignored by your eibc client
instance.
`,
		Args: cobra.ExactArgs(2),
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
			rollAppID := args[0]
			value := args[1]

			valueFloat, err := strconv.ParseFloat(value, 32)
			if err != nil {
				pterm.Error.Println("failed to convert value to float", err)
				return
			}

			updates := map[string]interface{}{
				fmt.Sprintf("fulfill_criteria.min_fee_percentage.chain.%s", rollAppID): valueFloat,
			}
			err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
			if err != nil {
				pterm.Error.Println("failed to update config", err)
				return
			}
		},
	}

	return cmd
}
