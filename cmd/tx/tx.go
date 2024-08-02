package tx

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/tx/claim"
	"github.com/dymensionxyz/roller/cmd/tx/fund_faucet"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Commands for sending transactions with Roller",
	}
	// cmd.AddCommand(register.Cmd())
	cmd.AddCommand(fund_faucet.Cmd())
	cmd.AddCommand(claim.Cmd())
	return cmd
}
