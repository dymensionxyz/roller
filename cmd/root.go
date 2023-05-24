package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "roller",
	Short: "A simple CLI tool to spin up a rollapp",
	Long: `
Roller CLI is a tool for registering and running autonomous rollapps built with Dymension RDK. The Roller offers everything you need to scaffold, config, register, and run your rollapp.
	`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
}
