package importkeys

import (
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "import <key-name> <priv-key-file-path>",
		Args:  cobra.ExactArgs(2),
		Short: "Exports the private key of the sequencer key.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			keyName := args[0]
			privKeyFilePath := args[1]
			keyringDir := filepath.Join(home, consts.ConfigDirName.HubKeys)

			expKeyArgs := []string{
				"keys",
				"import",
				keyName,
				privKeyFilePath,
				"--keyring-backend",
				"test",
				"--keyring-dir",
				keyringDir,
			}

			err := bash.ExecCommandWithInteractions(consts.Executables.Dymension, expKeyArgs...)
			if err != nil {
				pterm.Error.Println("failed to export private key: ", err)
				return
			}

			pterm.Info.Printf(
				"keys were imported into the test keyring, keyring-dir: %s",
				keyringDir,
			)
		},
	}

	return cmd
}
