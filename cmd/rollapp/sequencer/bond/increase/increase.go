package increase

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/tx"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "increase <amount>",
		Example: "roller rollapp sequencer bond increase 100000000000000000000adym",
		Short:   "Commands to manage sequencer instance",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()

			var amount string
			if len(args) != 0 {
				amount = args[0]
			} else {
				amount, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"please provide amount to add to the current bond:",
				).Show()

				if !strings.HasPrefix(amount, "adym") {
					pterm.Error.Println("invalid denom, only 'adym' is supported")
					return
				}
			}

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			c := exec.Command(
				consts.Executables.Dymension, "tx",
				"sequencer", "increase-bond", amount, "--keyring-backend",
				"test", "--from", "hub_sequencer", "--keyring-dir", filepath.Join(
					home,
					consts.ConfigDirName.HubKeys,
				), "--fees", fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
				"--node", rollerData.HubData.RPC_URL, "--chain-id", rollerData.HubData.ID,
			)

			txOutput, err := bash.ExecCommandWithInput(c, "signatures")
			if err != nil {
				pterm.Error.Println("failed to update bond: ", err)
				return
			}

			txHash, err := bash.ExtractTxHash(txOutput)
			if err != nil {
				pterm.Error.Println("failed to update bond: ", err)
				return
			}

			err = tx.MonitorTransaction(rollerData.HubData.RPC_URL, txHash)
			if err != nil {
				pterm.Error.Println("failed to update bond: ", err)
				return
			}
		},
	}

	return cmd
}
