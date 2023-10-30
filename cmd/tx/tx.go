package tx

import (
	"github.com/dymensionxyz/roller/cmd/tx/claim"
	"github.com/dymensionxyz/roller/cmd/tx/fund_faucet"
	"github.com/dymensionxyz/roller/cmd/tx/register"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tx",
		Short: "Commands for sending transactions with Roller",
	}
	cmd.AddCommand(register.Cmd())
	cmd.AddCommand(fund_faucet.Cmd())
	cmd.AddCommand(claim.Cmd())
	return cmd
}
