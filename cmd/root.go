package cmd

import (
	"os"

	da_light_client "github.com/dymensionxyz/roller/cmd/da-light-client"
	"github.com/spf13/cobra"

	blockexplorer "github.com/dymensionxyz/roller/cmd/block-explorer"
	"github.com/dymensionxyz/roller/cmd/eibc"
	"github.com/dymensionxyz/roller/cmd/relayer"
	"github.com/dymensionxyz/roller/cmd/rollapp"
	rollerutils "github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/cmd/version"
	"github.com/spf13/cobra"
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
	// rootCmd.AddCommand(config.Cmd())
	// rootCmd.AddCommand(version.Cmd())
	rootCmd.AddCommand(da_light_client.DALightClientCmd())
	rootCmd.AddCommand(relayer.Cmd())
	rootCmd.AddCommand(binaries.Cmd())
	// rootCmd.AddCommand(keys.Cmd())
	// rootCmd.AddCommand(run.Cmd())
	// rootCmd.AddCommand(services.Cmd())
	// rootCmd.AddCommand(migrate.Cmd())
	// rootCmd.AddCommand(tx.Cmd())
	// rootCmd.AddCommand(test())
	rootCmd.AddCommand(rollapp.Cmd())
	rootCmd.AddCommand(eibc.Cmd())
	rootCmd.AddCommand(blockexplorer.Cmd())
	rootCmd.AddCommand(version.Cmd())
	rollerutils.AddGlobalFlags(rootCmd)
}

func test() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test",
		Short: "Runs the rollapp on the local machine.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return nil
		},
	}
	return cmd
}
