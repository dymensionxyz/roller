package status

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/rollapp/start"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				fmt.Println("failed to load config:", err)
				return
			}
			errorhandling.PrettifyErrorIfExists(err)
			seq := sequencer.GetInstance(rollappConfig)
			fmt.Println(seq.GetSequencerStatus())

			pidFilePath := filepath.Join(home, consts.ConfigDirName.Rollapp, "pid")
			pid, err := os.ReadFile(pidFilePath)
			if err != nil {
				fmt.Println("failed to read pid file:", err)
				return
			}

			start.PrintOutput(rollappConfig, string(pid), true)
		},
	}
	return cmd
}
