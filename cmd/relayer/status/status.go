package status

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/logging"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the relayer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rollappConfig, err := roller.LoadConfig(home)

			relayerLogFilePath := logging.GetRelayerLogPath(home)
			errorhandling.PrettifyErrorIfExists(err)
			rly := relayer.NewRelayer(
				rollappConfig.Home,
				rollappConfig.RollappID,
				rollappConfig.HubData.ID,
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
