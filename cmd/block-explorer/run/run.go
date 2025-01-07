package run

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/blockexplorer"
	"github.com/dymensionxyz/roller/utils/config"
	relayerutils "github.com/dymensionxyz/roller/utils/relayer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a RollApp node.",
		Long:  ``,
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			var beRpcEndpoint string
			raID, _, _, err := relayerutils.GetRollappToRunFor(home, "block explorer")
			if err != nil {
				pterm.Error.Println("failed to determine what RollApp to run for:", err)
				return
			}

			for {
				// Prompt the user for the RPC URL
				beRpcEndpoint, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"rollapp block explorer rpc endpoint that you will provide (example: be.rollapp.dym.xyz), this endpoint should point to port 11100 of node",
				).Show()

				if strings.HasPrefix(beRpcEndpoint, "http://") {
					pterm.Error.Println("Invalid URL. Please try again.")
					continue // This will restart the loop from the beginning
				}

				if !strings.HasPrefix(beRpcEndpoint, "https://") {
					beRpcEndpoint = "https://" + beRpcEndpoint
				}

				isValid := config.IsValidURL(beRpcEndpoint)

				// Validate the URL
				if !isValid {
					pterm.Error.Println("Invalid URL. Please try again.")
					continue // This will also restart the loop
				}

				// Valid URL, break out of the loop
				break
			}

			hostAddress := "host.docker.internal"
			if runtime.GOOS == "linux" {
				hostAddress = "172.17.0.1" // Default Docker bridge network gateway
			}
			fmt.Println(hostAddress)

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

			err = blockexplorer.WriteChainsYAML(beChainConfigPath, beChainConfig)
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

	// cmd.Flags().
	// 	String("block-explorer-rpc-endpoint", "http://localhost:11100", "block explorer rpc endpoint")

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
