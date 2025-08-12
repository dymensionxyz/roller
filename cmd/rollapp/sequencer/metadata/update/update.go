package update

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/tx/tx_utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
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

			raData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			metadataFilePath := filepath.Join(
				home, consts.ConfigDirName.Rollapp, "init",
				"sequencer-metadata.json",
			)

			updateSeqArgs := []string{
				"tx",
				"sequencer",
				"update-sequencer",
				metadataFilePath,
				"--from",
				consts.KeysIds.HubSequencer,
				"--keyring-backend",
				string(raData.KeyringBackend),
				"--fees",
				fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
				"--gas-adjustment",
				"1.3",
				"--keyring-dir",
				filepath.Join(home, consts.ConfigDirName.HubKeys),
				"--node", raData.HubData.RpcUrl, "--chain-id", raData.HubData.ID,
			}
			var txHash string

			if raData.KeyringBackend == consts.SupportedKeyringBackends.OS {
				pswFileName, err := filesystem.GetOsKeyringPswFileName(consts.Executables.Dymension)
				if err != nil {
					pterm.Error.Println("failed to get os keyring psw file name", err)
					return
				}
				fp := filepath.Join(home, string(pswFileName))
				psw, err := filesystem.ReadFromFile(fp)
				if err != nil {
					pterm.Error.Println("failed to read keyring passphrase file", err)
					return
				}

				automaticPrompts := map[string]string{
					"Enter keyring passphrase":    psw,
					"Re-enter keyring passphrase": psw,
				}
				manualPromptResponses := map[string]string{
					"signatures": "this transaction is going to update the sequencer metadata. do you want to continue?",
				}

				txOutput, err := bash.ExecuteCommandWithPromptHandlerFiltered(
					consts.Executables.Dymension,
					updateSeqArgs,
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
				cmd := exec.Command(consts.Executables.Dymension, updateSeqArgs...)
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

			err = tx.MonitorTransaction(raData.HubData.WsUrl, txHash)
			if err != nil {
				pterm.Error.Println("transaction failed", err)
				return
			}

			var seqMetadata struct {
				GasPrice string `json:"gas_price"`
			}

			b, err := os.ReadFile(metadataFilePath)
			if err != nil {
				pterm.Error.Println("failed to read metadata file: ", err)
				return
			}

			err = json.Unmarshal(b, &seqMetadata)
			if err != nil {
				pterm.Error.Println("failed to unmarshal metadata file: ", err)
				return
			}

			appConfigFilePath := filepath.Join(
				sequencerutils.GetSequencerConfigDir(raData.Home),
				"app.toml",
			)
			pterm.Info.Println("setting minimum gas price in app.toml to", seqMetadata.GasPrice)
			err = tomlconfig.UpdateFieldInFile(
				appConfigFilePath,
				"minimum-gas-prices",
				seqMetadata.GasPrice,
			)
			if err != nil {
				pterm.Error.Println("failed to set minimum gas price in app.toml: ", err)
				return
			}
		},
	}

	return cmd
}
