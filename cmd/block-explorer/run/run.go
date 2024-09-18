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
	rollerfilesystemutils "github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
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

			hostAddress := "host.docker.internal"
			if runtime.GOOS == "linux" {
				hostAddress = "172.17.0.1" // Default Docker bridge network gateway
			}
			var raID string

			var beRpcEndpoint string
			if !isFlagChanged {
				useDefaultEndpoint, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
					WithDefaultText(
						fmt.Sprintf(
							`'--block-explorer-rpc-endpoint' is not set,
would you like to continue with the default endpoint (%s)?'
if you're running a rollapp locally, press 'y',
if you want to run block explorer for a rollapp on a different host, press 'n' and provide the endpoint and RollApp ID`,
							defaultBeRpcEndpoint,
						),
					).
					Show()
				if useDefaultEndpoint {
					beRpcEndpoint = fmt.Sprintf("http://%s:11100", hostAddress)

					err := rollerfilesystemutils.UpdateHostsFile(
						"127.0.0.1",
						"host.docker.internal",
					)
					if err != nil {
						pterm.Error.Println("failed to update hosts file", err)
						return
					}

					rollerData, err := tomlconfig.LoadRollerConfig(home)
					if err != nil {
						pterm.Error.Println("failed to load roller config file", err)
						return
					}

					raID = rollerData.RollappID
				} else {
					newBeRpcEndpoint, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
						"provide block explorer json rpc endpoint (running on port 11100 by default):",
					).Show()
					if newBeRpcEndpoint == "" {
						pterm.Error.Println("invalid endpoint")
						return
					}

					raIDInput, _ := pterm.DefaultInteractiveTextInput.WithDefaultText(
						"provide a rollapp ID that you want to run the node for",
					).Show()

					_, err := rollapp.ValidateChainID(raIDInput)
					if err != nil {
						pterm.Error.Println("invalid rollapp ID", err)
					}

					raID = raIDInput

					beRpcEndpoint = newBeRpcEndpoint
				}
			}

			beChainConfigPath := filepath.Join(
				home,
				consts.ConfigDirName.BlockExplorer,
				"config",
				"chains.yaml",
			)
			beChainConfig := blockexplorer.GenerateChainsYAML(
				raID,
				beRpcEndpoint,
			)
			err := blockexplorer.WriteChainsYAML(beChainConfigPath, beChainConfig)
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
