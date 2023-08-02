package show

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the channel data of the relayer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := config.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			rly := relayer.NewRelayer(rollappConfig.Home, rollappConfig.RollappID, rollappConfig.HubData.ID)
			srcChannel, dstChannel, err := rly.LoadChannels()
			utils.PrettifyErrorIfExists(err)
			if srcChannel == "" {
				fmt.Println("💈 No channel has been created for the relayer yet.")
			} else {
				fmt.Printf("💈 Relayer Channels: src, %s <-> %s, dst",
					srcChannel, dstChannel)
			}
		},
	}
	return cmd
}
