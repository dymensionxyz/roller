package show

import (
	"fmt"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show the channel data of the relayer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("show channel")
		},
	}
	return cmd
}
