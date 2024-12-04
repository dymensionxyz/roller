package set

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/eibc"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <rollapp-id> <comma-separated-full-nodes>",
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
			eibcConfigPath := filepath.Join(eibcHome, "config.yaml")
			var cfg eibcutils.Config
			err = cfg.LoadConfig(eibcConfigPath)
			if err != nil {
				pterm.Error.Println("failed to load eibc config: ", err)
				return
			}

			isEibcClientInitialized, err := filesystem.DirNotEmpty(eibcHome)
			if err != nil {
				pterm.Error.Println("failed to check eibc client initialized", err)
				return
			}

			if !isEibcClientInitialized {
				pterm.Error.Println("eibc client not initialized")
				return
			}

			rollAppID := args[0]
			fullNodes := args[1]
			if len(fullNodes) == 0 {
				pterm.Error.Println("please provide at least one full node")
				return
			}

			fNodes := strings.Split(fullNodes, ",")

			lspn, _ := pterm.DefaultSpinner.Start("adding rollapp to eibc config")
			err = eibc.AddRollappToEibcConfig(rollAppID, eibcHome, fNodes)
			if err != nil {
				return
			}
			lspn.Success("rollapp added to eibc config")

			rspn, _ := pterm.DefaultSpinner.Start("updating eibc operator metadata")
			rspn.UpdateText("retrieving updated supported rollapp list")
			raIDs, err := eibcutils.LoadSupportedRollapps(eibcConfigPath)
			if err != nil {
				pterm.Error.Println("failed to load supported rollapps: ", err)
				return
			}
			hd, err := cfg.HubDataFromHubRpc(eibcConfigPath)
			if err != nil {
				pterm.Error.Println("failed to retrieve hub data: ", err)
				return
			}

			rspn.UpdateText("retrieving existing eibc operator metadata")
			metadata, err := eibcutils.EibcOperatorMetadataFromChain(home, *hd)
			if err != nil {
				pterm.Error.Println("failed to retrieve eibc operator metadata: ", err)
				return
			}

			j, err := json.Marshal(metadata)
			if err != nil {
				pterm.Error.Println("failed to marshal eibc operator metadata: ", err)
				return
			}
			fmt.Println(string(j))

			rspn.UpdateText("updating supported rollapp list")
			metadata.SupportedRollapps = raIDs

			mb, err := metadata.ToBytes()
			if err != nil {
				pterm.Error.Println("failed to generate eibc operator metadata: ", err)
				return
			}
			mbs := base64.StdEncoding.EncodeToString(mb)

			rspn.UpdateText("pushing changes to chain")
			err = eibcutils.UpdateEibcOperatorMetadata(home, mbs, *hd)
			if err != nil {
				pterm.Error.Println("failed to update eibc operator metadata: ", err)
				return
			}
			rspn.Success("operator metadata updated, new metadata:")
			ym, err := yaml.Marshal(metadata)
			if err != nil {
				pterm.Error.Println("failed to marshal eibc operator metadata: ", err)
				return
			}
			fmt.Println(string(ym))
		},
	}

	return cmd
}
