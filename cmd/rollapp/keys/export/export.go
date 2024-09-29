package export

import (
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Exports the private key of the sequencer key.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollerData, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)

			var kcs []utils.KeyConfig
			if rollerData.HubData.ID != "mock" {
				kcs = initconfig.GetSequencerKeysConfig()
			} else {
				kcs = initconfig.GetMockSequencerKeyConfig(rollerData)
			}

			kc := kcs[0]

			expKeyArgs := []string{
				"keys",
				"export",
				kc.ID,
				"--keyring-backend",
				"test",
				"--keyring-dir",
				filepath.Join(home, consts.ConfigDirName.HubKeys),
			}

			err = bash.ExecCommandWithInteractions(kc.ChainBinary, expKeyArgs...)
			if err != nil {
				pterm.Error.Println("failed to export private key: ", err)
				return
			}
		},
	}

	return cmd
}
