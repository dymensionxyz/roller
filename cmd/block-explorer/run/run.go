package run

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Run a RollApp node.",
		Long:  ``,
		Args:  cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Run the block explorer")

			err := createBlockExplorerContainers()
			if err != nil {
				pterm.Error.Println("failed to create the necessary containers: ", err)
				return
			}
		},
	}

	return cmd
}
