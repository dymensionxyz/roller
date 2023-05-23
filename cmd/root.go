/*
Copyright © 2023 Dymension itay@dymension.xyz

*/
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
	-h, --help   help for ignite

Use "roller [command] --help" for more information about a command.
	`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.roller.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
