package tx

import (
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Commands for sending transactions with Roller",
	}
	// cmd.AddCommand(register.Cmd())
	// cmd.AddCommand(fund_faucet.Cmd())
	// cmd.AddCommand(claim.Cmd())
	return cmd
}
