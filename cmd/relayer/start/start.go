package start

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/spf13/cobra"
)

func start() *cobra.Command {
	registerCmd := &cobra.Command{

		Use:   "start",
		Short: "Starts a relayer between the Dymension hub and the rollapp.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollappConfig, err := utils.LoadConfigFromTOML(home)
			utils.PrettifyErrorIfExists(err)
			fmt.Println(rollappConfig)
			createIBCChannel()

		},
	}
	utils.AddGlobalFlags(registerCmd)
	return registerCmd
}

func createIBCChannel() {
	fmt.Println("Creating IBC channel...")
}
