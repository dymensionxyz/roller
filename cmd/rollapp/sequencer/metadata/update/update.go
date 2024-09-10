package update

import (
	"fmt"
	"os/exec"
	"path/filepath"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/tx"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
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

			home, err := globalutils.ExpandHomePath(cmd.Flag(utils.FlagNames.Home).Value.String())
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}

			rollerData, err := tomlconfig.LoadRollerConfig(home)
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
				filepath.Join(utils.GetRollerRootDir(), consts.ConfigDirName.HubKeys),
				"--node", rollerData.HubData.RPC_URL, "--chain-id", rollerData.HubData.ID,
			)

			txOutput, err := bash.ExecCommandWithInput(updateSeqCmd, "signatures")
			if err != nil {
				pterm.Error.Println("failed to update sequencer metadata", err)
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
