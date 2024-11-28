package status

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/logging"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the relayer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rlyConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.Relayer,
				"config",
				"config.yaml",
			)
			relayerLogFilePath := logging.GetRelayerLogPath(home)

			var rlyCfg relayer.Config
			err := rlyCfg.Load(rlyConfigPath)
			if err != nil {
				pterm.Error.Println("failed to load relayer config: ", err)
				return
			}

			raData := rlyCfg.RaDataFromRelayerConfig()
			hd := rlyCfg.HubDataFromRelayerConfig()

			rly := relayer.NewRelayer(
				home,
				*raData,
				*hd,
			)

			bytes, err := os.ReadFile(rly.StatusFilePath())
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fmt.Println("💈 Starting...")
					return
				}
			} else {
				errorhandling.PrettifyErrorIfExists(err)
			}
			fmt.Println(string(bytes))
			fmt.Println("💈 Log file path: ", relayerLogFilePath)
		},
	}
	return cmd
}
