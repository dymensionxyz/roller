package unbond

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/tx"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unbond",
		Short: "Commands to manage sequencer instance",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			c := exec.Command(
				consts.Executables.Dymension, "tx",
				"sequencer", "unbond", "--keyring-backend",
				"test", "--from", consts.KeysIds.HubSequencer, "--keyring-dir", filepath.Join(
					home,
					consts.ConfigDirName.HubKeys,
				), "--fees", fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
				"--node", rollerData.HubData.RpcUrl, "--chain-id", rollerData.HubData.ID,
			)

			txOutput, err := bash.ExecCommandWithInput(home, c, "signatures")
			if err != nil {
				pterm.Error.Println("failed to update bond: ", err)
				return
			}

			txHash, err := bash.ExtractTxHash(txOutput)
			if err != nil {
				pterm.Error.Println("failed to update bond: ", err)
				return
			}

			if rollerData.HubData.WsUrl == "" {
				err = tx.MonitorTransaction(rollerData.HubData.RpcUrl, txHash)
			} else {
				err = tx.MonitorTransaction(rollerData.HubData.WsUrl, txHash)
			}
			if err != nil {
				pterm.Error.Println("failed to update bond: ", err)
				return
			}
		},
	}

	return cmd
}
