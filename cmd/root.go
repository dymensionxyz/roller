package cmd

import (
	"os"

	"github.com/spf13/cobra"

	blockexplorer "github.com/dymensionxyz/roller/cmd/block-explorer"
	"github.com/dymensionxyz/roller/cmd/config"
	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	da_light_client "github.com/dymensionxyz/roller/cmd/da-light-client"
	"github.com/dymensionxyz/roller/cmd/eibc"
	"github.com/dymensionxyz/roller/cmd/observability"
	"github.com/dymensionxyz/roller/cmd/relayer"
	"github.com/dymensionxyz/roller/cmd/rollapp"
	"github.com/dymensionxyz/roller/cmd/rollapp/keys"
	"github.com/dymensionxyz/roller/cmd/version"
)

var rootCmd = &cobra.Command{
	Use:   "roller",
	Short: "A simple CLI tool to spin up a RollApp",
	Long: `
Roller CLI is a tool for registering and running autonomous RollApps built with Dymension RDK. Roller provides everything you need to scaffold, configure, register, and run your RollApp.
	`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(da_light_client.DALightClientCmd())
	rootCmd.AddCommand(relayer.Cmd())
	rootCmd.AddCommand(keys.Cmd())
	rootCmd.AddCommand(observability.Cmd())
	rootCmd.AddCommand(rollapp.Cmd())
	rootCmd.AddCommand(eibc.Cmd())
	rootCmd.AddCommand(blockexplorer.Cmd())
	rootCmd.AddCommand(config.Cmd())
	rootCmd.AddCommand(version.Cmd())

	initconfig.AddGlobalFlags(rootCmd)
}
