package decrease

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/tx/tx_utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/tx"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "decrease <amount>",
		Example: "roller rollapp sequencer bond increase 100000000000000000000adym",
		Short:   "Commands to increase the sequencer bond amount",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			var amount string
			if len(args) != 0 {
				amount = args[0]
			} else {
				amount, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
					"please provide amount to remove from the current bond:",
				).Show()

				if !strings.HasSuffix(amount, "adym") {
					pterm.Error.Println("invalid denom, only 'adym' is supported")
					return
				}
			}

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			bondArgs := []string{
				"tx",
				"sequencer",
				"decrease-bond",
				amount,
				"--keyring-backend",
				string(rollerData.KeyringBackend),
				"--from",
				consts.KeysIds.HubSequencer,
				"--keyring-dir",
				filepath.Join(
					home,
					consts.ConfigDirName.HubKeys,
				),
				"--fees",
				fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
				"--node",
				rollerData.HubData.RpcUrl,
				"--chain-id",
				rollerData.HubData.ID,
			}
			var txHash string

			if rollerData.KeyringBackend == consts.SupportedKeyringBackends.OS {
				psw, err := filesystem.ReadOsKeyringPswFile(home, consts.Executables.Dymension)
				if err != nil {
					pterm.Error.Println("failed to read os keyring password file", err)
					return
				}

				automaticPrompts := map[string]string{
					"Enter keyring passphrase":    psw,
					"Re-enter keyring passphrase": psw,
				}
				manualPromptResponses := map[string]string{
					"signatures": "this transaction is going to update the sequencer metadata. do you want to continue?",
				}

				txOutput, err := bash.ExecuteCommandWithPromptHandler(
					consts.Executables.Dymension,
					bondArgs,
					automaticPrompts,
					manualPromptResponses,
				)
				if err != nil {
					pterm.Error.Println("failed to update sequencer metadata", err)
					return
				}
				tob := bytes.NewBufferString(txOutput.String())
				err = tx_utils.CheckTxYamlStdOut(*tob)
				if err != nil {
					pterm.Error.Println("failed to check raw_log", err)
					return
				}

				txHash, err = bash.ExtractTxHash(txOutput.String())
				if err != nil {
					pterm.Error.Println("failed to extract tx hash", err)
					return
				}
			} else {
				cmd := exec.Command(consts.Executables.Dymension, bondArgs...)
				txOutput, err := bash.ExecCommandWithInput(home, cmd, "signatures")
				if err != nil {
					pterm.Error.Println("failed to update sequencer metadata", err)
					return
				}

				txHash, err = bash.ExtractTxHash(txOutput)
				if err != nil {
					pterm.Error.Println("failed to extract tx hash", err)
					return
				}
			}

			err = tx.MonitorTransaction(rollerData.HubData.RpcUrl, txHash)
			if err != nil {
				pterm.Error.Println("failed to update bond: ", err)
				return
			}
		},
	}

	return cmd
}
