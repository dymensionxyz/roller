package export

import (
	"fmt"
	"path/filepath"
	"strings"

	datalayer "github.com/dymensionxyz/roller/data_layer"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <key-id>",
		Short: "Exports the private key of the given key id.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			var supportedKeys = []string{
				consts.KeysIds.HubSequencer,
				consts.KeysIds.RollappSequencer,
			}
			damanager := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
			if damanager.GetKeyName() != "" {
				supportedKeys = append(supportedKeys, damanager.GetKeyName())
			}
			keyID := args[0]
			if keyID == consts.KeysIds.HubSequencer {
				exportKeyCmd := utils.GetExportKeyCmdBinary(keyID, filepath.Join(home, consts.ConfigDirName.HubKeys),
					consts.Executables.Dymension)
				out, err := utils.ExecBashCommand(exportKeyCmd)
				printKeyOutput(out.String(), err)
			} else if keyID == consts.KeysIds.RollappSequencer {
				exportKeyCmd := utils.GetExportKeyCmdBinary(keyID, filepath.Join(home, consts.ConfigDirName.Rollapp),
					rlpCfg.RollappBinary)
				out, err := utils.ExecBashCommand(exportKeyCmd)
				printKeyOutput(out.String(), err)
			} else if keyID != "" && keyID == damanager.GetKeyName() {
				//TODO: avail doesn't need cmd to get the keys, it's stored in the config
				exportKeyCmd := damanager.GetExportKeyCmd()
				// TODO: make more generic. need it because cel-key write the output to stderr for some reason
				if exportKeyCmd != nil {
					out, err := utils.ExecBashCommandWithStdErr(exportKeyCmd)
					printKeyOutput(out.String(), err)
				}
			} else {
				utils.PrettifyErrorIfExists(fmt.Errorf("invalid key id: %s. The supported keys are %s", keyID,
					strings.Join(supportedKeys, ", ")))
			}
		},
		Args: cobra.ExactArgs(1),
	}

	return cmd
}

func printKeyOutput(output string, err error) {
	utils.PrettifyErrorIfExists(err)
	fmt.Printf("🔑 Unarmored Hex Private Key: %s", output)
}
