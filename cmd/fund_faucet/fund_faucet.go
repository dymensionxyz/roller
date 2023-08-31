package fund_faucet

import (
	"fmt"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "fund-faucet",
		Short: "Fund the Dymension faucet with rollapp tokens",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("funding it")
		},
	}
	return versionCmd
}
