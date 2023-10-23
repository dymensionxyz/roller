package claim

import (
	"fmt"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "claim-rewards <private-key> <destination address>",
		Short: "Send the DYM rewards associated with the given private key to the destination address",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello World")
		},
	}
	return cmd
}
