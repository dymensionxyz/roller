package cmd

import (
	"os"

	"github.com/dymensionxyz/roller/cmd/config"
	"github.com/dymensionxyz/roller/cmd/register"
	"github.com/dymensionxyz/roller/cmd/run"
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
	rootCmd.AddCommand(config.ConfigCmd())
	rootCmd.AddCommand(version.VersionCmd())
	rootCmd.AddCommand(register.RegisterCmd())
	rootCmd.AddCommand(run.RunCmd())
}
