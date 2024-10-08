package update

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"

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
		Use:   "update [metadata-file.json]",
		Short: "Update the sequencer metadata",
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

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			metadataFilePath := filepath.Join(
				home, consts.ConfigDirName.Rollapp, "init",
				"sequencer-metadata.json",
			)

			updateSeqCmd := exec.Command(
				consts.Executables.Dymension,
				"tx",
				"sequencer",
				"update-sequencer",
				metadataFilePath,
				"--from",
				consts.KeysIds.HubSequencer,
				"--keyring-backend",
				"test",
				"--fees",
				fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
				"--gas-adjustment",
				"1.3",
				"--keyring-dir",
				filepath.Join(roller.GetRootDir(), consts.ConfigDirName.HubKeys),
				"--node", rollerData.HubData.RPC_URL, "--chain-id", rollerData.HubData.ID,
			)

			txOutput, err := bash.ExecCommandWithInput(updateSeqCmd, "signatures")
			if err != nil {
				pterm.Error.Println("failed to update sequencer metadata", err)
				return
			}

			tob := bytes.NewBufferString(txOutput)
			err = tx_utils.CheckTxYamlStdOut(*tob)
			if err != nil {
				pterm.Error.Println("failed to check raw_log", err)
				return
			}

			txHash, err := bash.ExtractTxHash(txOutput)
			if err != nil {
				return
			}

			err = tx.MonitorTransaction(rollerData.HubData.RPC_URL, txHash)
			if err != nil {
				pterm.Error.Println("transaction failed", err)
				return
			}
		},
	}

	return cmd
}
