package cmd

import (
	"github.com/spf13/cobra"
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Commands for setting up and managing rollapp configuration files.",
}

func init() {
	rootCmd.AddCommand(configCmd)
}
