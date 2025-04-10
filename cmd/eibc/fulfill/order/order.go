package order

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/eibc"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
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

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerCfg, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config: ", err)
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

			var feeAmount string
			if len(args) != 0 {
				feeAmount = args[1]
			} else {
				feeAmount, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"provide the expected fee amount",
				).Show()
			}

			gCmd, err := eibc.GetFulfillOrderCmd(
				orderId,
				feeAmount,
				rollerCfg.HubData,
			)
			if err != nil {
				pterm.Error.Println("failed to fulfill order: ", err)
				return
			}

			txOutput, err := bash.ExecCommandWithInput(home, gCmd, "signatures")
			if err != nil {
				pterm.Error.Println("failed to update sequencer metadata", err)
				return
			}

			txHash, err := bash.ExtractTxHash(txOutput)
			if err != nil {
				pterm.Error.Println("failed to extract tx hash", err)
				return
			}

			err = tx.MonitorTransaction(rollerCfg.HubData.WsUrl, txHash)
			if err != nil {
				pterm.Error.Println("transaction failed", err)
				return
			}
		},
	}

	return cmd
}
