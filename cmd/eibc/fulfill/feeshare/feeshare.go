package feeshare

import (
	"os"
	"path/filepath"
	"strconv"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "fee-share [fee-share]",
		Short:   "Set",
		Example: "roller eibc fulfill fee-share 0.1",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			home, err := os.UserHomeDir()
			if err != nil {
				pterm.Error.Println("failed to get user home dir", err)
				return
			}

			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			eibcHome := filepath.Join(home, consts.ConfigDirName.Eibc)
			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")

			fee := args[0]
			ff, err := strconv.ParseFloat(fee, 64)
			if err != nil {
				pterm.Error.Printf("fee must be a valid number, got %s\n", fee)
				return
			}

			updates := map[string]interface{}{
				"operator.min_fee_share": ff,
			}
			err = yamlconfig.UpdateNestedYAML(eibcConfigPath, updates)
			if err != nil {
				pterm.Error.Println("failed to update config", err)
				return
			}

			var cfg eibcutils.Config
			err = cfg.LoadConfig(eibcConfigPath)
			if err != nil {
				pterm.Error.Println("failed to load eibc config: ", err)
				return
			}

			err = eibcutils.UpdateGroupOperatorMinFee(eibcConfigPath, ff, cfg, home)
			if err != nil {
				pterm.Error.Println("failed to update eibc operator metadata: ", err)
				return
			}
		},
	}

	return cmd
}
