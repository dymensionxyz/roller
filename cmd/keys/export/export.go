package export

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/toml"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export <key-id>",
		Short: "Exports the private key of the given key id.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := toml.LoadRollerConfigFromTOML(home)
			errorhandling.PrettifyErrorIfExists(err)
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
				out, err := bash.ExecCommandWithStdout(exportKeyCmd)
				errorhandling.PrettifyErrorIfExists(err)
				printHexKeyOutput(out.String())
			} else if keyID == consts.KeysIds.RollappSequencer {
				exportKeyCmd := utils.GetExportKeyCmdBinary(
					keyID, filepath.Join(home, consts.ConfigDirName.Rollapp),
					rlpCfg.RollappBinary,
				)
				out, err := bash.ExecCommandWithStdout(exportKeyCmd)
				errorhandling.PrettifyErrorIfExists(err)
				printHexKeyOutput(out.String())
			} else if keyID != "" && keyID == damanager.GetKeyName() {
				privateKey, err := damanager.GetPrivateKey()
				errorhandling.PrettifyErrorIfExists(err)
				if rlpCfg.DA == consts.Celestia {
					printHexKeyOutput(privateKey)
				} else if rlpCfg.DA == consts.Avail {
					printMnemonicKeyOutput(privateKey)
				}
			} else {
				errorhandling.PrettifyErrorIfExists(
					fmt.Errorf(
						"invalid key id: %s. The supported keys are %s", keyID,
						strings.Join(supportedKeys, ", "),
					),
				)
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
