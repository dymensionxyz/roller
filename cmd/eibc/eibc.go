package eibc

import (
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "eibc",
		Short: "Commands for running and managing eibc client",
	}

	cmd.AddCommand(initCmd())
	cmd.AddCommand(startCmd())

	return cmd
}
