package export

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <key-id>",
		Short: "Exports the private key of the given key id.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			supportedKeys := []string{
				consts.KeysIds.HubSequencer,
				consts.KeysIds.RollappSequencer,
			}
			damanager := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
			if damanager.GetKeyName() != "" {
				supportedKeys = append(supportedKeys, damanager.GetKeyName())
			}
			keyID := args[0]
			if keyID == consts.KeysIds.HubSequencer {
				exportKeyCmd := utils.GetExportKeyCmdBinary(
					keyID,
					filepath.Join(home, consts.ConfigDirName.HubKeys),
					consts.Executables.Dymension,
				)
				out, err := utils.ExecBashCommandWithStdout(exportKeyCmd)
				utils.PrettifyErrorIfExists(err)
				printHexKeyOutput(out.String())
			} else if keyID == consts.KeysIds.RollappSequencer {
				exportKeyCmd := utils.GetExportKeyCmdBinary(keyID, filepath.Join(home, consts.ConfigDirName.Rollapp),
					rlpCfg.RollappBinary)
				out, err := utils.ExecBashCommandWithStdout(exportKeyCmd)
				utils.PrettifyErrorIfExists(err)
				printHexKeyOutput(out.String())
			} else if keyID != "" && keyID == damanager.GetKeyName() {
				privateKey, err := damanager.GetPrivateKey()
				utils.PrettifyErrorIfExists(err)
				if rlpCfg.DA == config.Celestia {
					printHexKeyOutput(privateKey)
				} else if rlpCfg.DA == config.Avail {
					printMnemonicKeyOutput(privateKey)
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

func printMnemonicKeyOutput(key string) {
	fmt.Printf("🔑 Mnemonic: %s\n", key)
}

func printHexKeyOutput(key string) {
	fmt.Printf("🔑 Unarmored Hex Private Key: %s", key)
}
