package status

import (
	"errors"
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/spf13/cobra"
	"os"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show the status of the relayer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			rly := relayer.NewRelayer(rollappConfig.Home, rollappConfig.RollappID, rollappConfig.HubData.ID)

			bytes, err := os.ReadFile(rly.StatusFilePath())
			if err != nil {
				if errors.Is(err, os.ErrNotExist) {
					fmt.Println("💈 Starting...")
					return
				}
			} else {
				utils.PrettifyErrorIfExists(err)
			}
			fmt.Println(string(bytes))
		},
	}
	return cmd
}
