package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "roller",
	Short: "A simple CLI tool to spin up a rollapp",
	Long: `
Roller CLI is a tool for registering and running autonomous rollapps built with Dymension RDK. Dymension CLI offers everything you need to scaffold, config, register, and run your rollapp.

To get started, init a rollapp configuration on your local machine:

roller init mars

Usage:
	roller [command]

Available Commands:
	scaffold    Scaffold a new boilerplate dymension-rdk rollapp
	config      Commands to initialize and manage the rollapp config on your local machine
	register    Register a rollapp with the Dymension RDK
	run         Run a rollapp locally
	connect     Connect a runnning rollapp to the Dymension hub via IBC
	version     Print the current build information
	help        Help about any command
	completion  Generate the autocompletion script for the specified shell

Flags:
	-h, --help   help for roller

Use "roller [command] --help" for more information about a command.
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
