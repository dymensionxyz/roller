package binaries

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/binaries/versions"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "binaries",
		Short: "Commands to manage roller dependencies",
	}

	// cmd.AddCommand(install.Cmd())
	cmd.AddCommand(versions.Cmd())

	return cmd
}
