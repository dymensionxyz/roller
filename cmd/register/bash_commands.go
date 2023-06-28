package register

import (
	"github.com/dymensionxyz/roller/cmd/utils"
	"os/exec"
	"path/filepath"

	"fmt"

	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func getRegisterRollappCmd(rollappConfig utils.RollappConfig) *exec.Cmd {
	cmdArgs := []string{
		"tx", "rollapp", "create-rollapp", rollappConfig.RollappID, "stamp1", "genesis-path/1", "3", "3", `{"Addresses":[]}`,
	}
	cmdArgs = append(cmdArgs, getCommonDymdTxFlags(rollappConfig)...)
	return exec.Command(
		consts.Executables.Dymension, cmdArgs...,
	)
}

func showSequencerPubKey(rollappConfig utils.RollappConfig) (string, error) {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"dymint",
		"show-sequencer",
		"--home",
		filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
	)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.ReplaceAll(strings.ReplaceAll(string(out), "\n", ""), "\\", ""), nil
}

func getRegisterSequencerCmd(rollappConfig utils.RollappConfig) (*exec.Cmd, error) {
	seqPubKey, err := showSequencerPubKey(rollappConfig)
	if err != nil {
		return nil, err
	}
	description := fmt.Sprintf(`{"Moniker":"%s","Identity":"","Website":"","SecurityContact":"","Details":""}`,
		consts.KeysIds.HubSequencer)
	cmdArgs := []string{
		"tx", "sequencer", "create-sequencer",
		seqPubKey,
		rollappConfig.RollappID,
		description,
	}
	cmdArgs = append(cmdArgs, getCommonDymdTxFlags(rollappConfig)...)
	return exec.Command(consts.Executables.Dymension, cmdArgs...), nil
}

func getCommonDymdTxFlags(rollappConfig utils.RollappConfig) []string {
	commonFlags := utils.GetCommonDymdFlags(rollappConfig)
	txArgs := []string{
		"--from", consts.KeysIds.HubSequencer,
		"--keyring-backend", "test",
		"--keyring-dir", filepath.Join(rollappConfig.Home, consts.ConfigDirName.HubKeys),
		"--yes", "--broadcast-mode", "block", "--chain-id", rollappConfig.HubData.ID,
	}
	return append(commonFlags, txArgs...)
}
