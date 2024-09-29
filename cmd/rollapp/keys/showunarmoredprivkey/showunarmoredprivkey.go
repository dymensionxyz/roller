package showunarmoredprivkey

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-unarmored-priv-key",
		Short: "Exports the private key of the sequencer key.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rlpCfg, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			var exportKeyCmd *exec.Cmd
			var keyID string

			if rlpCfg.HubData.ID != "mock" {
				keyID = consts.KeysIds.HubSequencer
			} else {
				keyID = consts.KeysIds.RollappSequencer
			}

			exportKeyCmd = utils.GetExportKeyCmdBinary(
				keyID,
				filepath.Join(home, consts.ConfigDirName.HubKeys),
				consts.Executables.Dymension,
			)

			out, err := bash.ExecCommandWithStdout(exportKeyCmd)
			errorhandling.PrettifyErrorIfExists(err)

			pterm.DefaultSection.WithIndentCharacter("🔑").Println(keyID)
			printHexKeyOutput(out.String())
		},
	}

	return cmd
}

// nolint: unused
func printMnemonicKeyOutput(key string) {
	fmt.Printf("🔑 Mnemonic: %s\n", key)
}

func printHexKeyOutput(key string) {
	fmt.Printf("Unarmored Hex Private Key: %s", key)
}
