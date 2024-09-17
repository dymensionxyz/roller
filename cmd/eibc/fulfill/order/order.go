package order

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/tx"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "order",
		Short: "Commands related to fulfillment of eibc orders",
		Args:  cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			err := initconfig.AddFlags(cmd)
			if err != nil {
				pterm.Error.Println("failed to add flags")
				return
			}

			home, err := filesystem.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerCfg, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				return
			}

			var orderId string
			if len(args) != 0 {
				orderId = args[0]
			} else {
				orderId, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide an order id that you want to fulfill",
				).Show()
			}

			var percentage string
			if len(args) != 0 {
				percentage = args[1]
			} else {
				percentage, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide the expected fee amount",
				).Show()
			}

			fmt.Println("orders related to fulfillment of eibc orders")
			gCmd, err := eibc.GetFulfillOrderCmd(
				orderId,
				percentage,
				rollerCfg.HubData,
			)
			if err != nil {
				pterm.Error.Println("failed to fulfill order: ", err)
				return
			}

			txOutput, err := bash.ExecCommandWithInput(gCmd, "signatures")
			if err != nil {
				pterm.Error.Println("failed to update sequencer metadata", err)
				return
			}

			txHash, err := bash.ExtractTxHash(txOutput)
			if err != nil {
				return
			}

			err = tx.MonitorTransaction(rollerCfg.HubData.RPC_URL, txHash)
			if err != nil {
				pterm.Error.Println("transaction failed", err)
				return
			}
		},
	}

	return cmd
}
