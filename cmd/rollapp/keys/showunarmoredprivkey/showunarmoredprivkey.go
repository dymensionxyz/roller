package showunarmoredprivkey

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show-unarmored-priv-key",
		Short: "Exports the private key of the sequencer key.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rlpCfg, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			var exportKeyCmd *exec.Cmd
			var keyID string

			if rlpCfg.HubData.ID != "mock" {
				keyID = consts.KeysIds.HubSequencer
			} else {
				keyID = consts.KeysIds.RollappSequencer
			}

			exportKeyCmd = keys.GetExportKeyCmdBinary(
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
