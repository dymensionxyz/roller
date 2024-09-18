package run

import (
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/blockexplorer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a RollApp node.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			isFlagChanged := cmd.Flags().Changed("block-explorer-rpc-endpoint")
			defaultBeRpcEndpoint, _ := cmd.Flags().GetString("block-explorer-rpc-endpoint")

			var beRpcEndpoint string
			if !isFlagChanged {
				useDefaultEndpoint, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
					WithDefaultText(
						fmt.Sprintf(
							`'--block-explorer-rpc-endpoint' is not set,
would you like to continue with the default endpoint (%s)?'
press 'y' if you're running local node, press 'n' and provide the endpoint if you're running remote node`,
							defaultBeRpcEndpoint,
						),
					).
					Show()
				if useDefaultEndpoint {
					beRpcEndpoint = defaultBeRpcEndpoint
				} else {
					newBeRpcEndpoint, _ := pterm.DefaultInteractiveTextInput.Show()
					beRpcEndpoint = newBeRpcEndpoint
				}
			}

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			beChainConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.BlockExplorer,
				"config",
				"chains.yaml",
			)
			beChainConfig := blockexplorer.GenerateChainsYAML(
				rollerData.RollappID,
				beRpcEndpoint,
			)
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

	cmd.Flags().
		String("block-explorer-rpc-endpoint", "http://localhost:11100", "block explorer rpc endpoint")

	return cmd
}
