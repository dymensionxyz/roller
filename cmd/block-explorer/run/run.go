package run

import (
	"fmt"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/blockexplorer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a RollApp node.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			fmt.Println("Run the block explorer")

			beChainConfigPath := filepath.Join(home, "block-explorer", "config", "chains.yaml")
			beChainConfig := blockexplorer.GenerateChainsYAML(rollerData.RollappID)
			err = blockexplorer.WriteChainsYAML(beChainConfigPath, beChainConfig)
			if err != nil {
				pterm.Error.Println("failed to generate block explorer config", err)
			}

			err = createBlockExplorerContainers(home)
			if err != nil {
				pterm.Error.Println("failed to create the necessary containers: ", err)
				return
			}
		},
	}

	return cmd
}
