package run

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/blockexplorer"
	rollerfilesystemutils "github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a RollApp node.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			isFlagChanged := cmd.Flags().Changed("block-explorer-rpc-endpoint")
			defaultBeRpcEndpoint, _ := cmd.Flags().GetString("block-explorer-rpc-endpoint")

			hostAddress := "host.docker.internal"
			if runtime.GOOS == "linux" {
				// Try to get the Docker bridge network gateway dynamically
				output, err := exec.Command("docker", "network", "inspect", "bridge", "-f", "{{range .IPAM.Config}}{{.Gateway}}{{end}}").
					Output()
				if err != nil {
					hostAddress = "172.17.0.1" // Fallback to default gateway
				} else {
					hostAddress = strings.TrimSpace(string(output))
				}

				// Ensure host.docker.internal is available on Linux
				err = rollerfilesystemutils.UpdateHostsFile(hostAddress, "host.docker.internal")
				if err != nil {
					pterm.Warning.Printf(
						"Failed to update hosts file: %v. Using IP address directly.\n",
						err,
					)
				}
			}
			var raID string
			fmt.Println(hostAddress)

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

					rollerData, err := roller.LoadConfig(home)
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

			fmt.Println(beChainConfig)

			err := blockexplorer.WriteChainsYAML(beChainConfigPath, beChainConfig)
			if err != nil {
				pterm.Error.Println("failed to generate block explorer config", err)
			}

			err = createBlockExplorerContainers(home, beRpcEndpoint)
			if err != nil {
				pterm.Error.Println("failed to create the necessary containers: ", err)
				return
			}

			printOutput(raID, beRpcEndpoint)
		},
	}

	cmd.Flags().
		String("block-explorer-rpc-endpoint", "http://localhost:11100", "block explorer rpc endpoint")

	return cmd
}

func printOutput(raID, beRpcEndpoint string) {
	pterm.DefaultBasicText.WithStyle(
		pterm.
			FgGreen.ToStyle(),
	).Sprintf("💈 RollApp Block Explorer is running locally")

	pterm.DefaultSection.WithIndentCharacter("💈").
		Println("Endpoints:")
	fmt.Println("Block Explorer: http://localhost:3000")

	pterm.DefaultSection.WithIndentCharacter("💈").
		Println("RollApp Information:")
	fmt.Println("RollApp ID: ", raID)
	fmt.Println("Block Explorer API Endpoint: ", beRpcEndpoint)

	pterm.DefaultSection.WithIndentCharacter("💈").
		Println("Container Information:")
	fmt.Println("Block Explorer: ", "be-frontend")
	fmt.Println("Indexer: ", "be-indexer")
	fmt.Println("Database: ", "be-postgresql")
}
