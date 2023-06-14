package register

import (
	"github.com/dymensionxyz/roller/cmd/utils"
	"os/exec"
	"path/filepath"

	"fmt"

	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func getRegisterRollappCmd(rollappConfig utils.InitConfig) *exec.Cmd {
	cmdArgs := []string{
		"tx", "rollapp", "create-rollapp", rollappConfig.RollappID, "stamp1", "genesis-path/1", "3", "3", `{"Addresses":[]}`,
	}
	cmdArgs = append(cmdArgs, GetCommonFlags(rollappConfig)...)
	return exec.Command(
		consts.Executables.Dymension, cmdArgs...,
	)
}

func showSequencerPubKey(rollappConfig utils.InitConfig) (string, error) {
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

func getRegisterSequencerCmd(rollappConfig utils.InitConfig) (*exec.Cmd, error) {
	seqPubKey, err := showSequencerPubKey(rollappConfig)
	if err != nil {
		return nil, err
	}
	description := fmt.Sprintf(`{"Moniker":"%s","Identity":"","Website":"","SecurityContact":"","Details":""}`,
		consts.KeyNames.HubSequencer)
	cmdArgs := []string{
		"tx", "sequencer", "create-sequencer",
		seqPubKey,
		rollappConfig.RollappID,
		description,
	}
	cmdArgs = append(cmdArgs, GetCommonFlags(rollappConfig)...)
	return exec.Command(consts.Executables.Dymension, cmdArgs...), nil
}

func GetCommonFlags(rollappConfig utils.InitConfig) []string {
	return []string{
		"--from", consts.KeyNames.HubSequencer,
		"--keyring-backend", "test",
		"--keyring-dir", filepath.Join(rollappConfig.Home, consts.ConfigDirName.Rollapp),
		"--node", rollappConfig.HubData.RPC_URL, "--output", "json",
		"--yes", "--broadcast-mode", "block", "--chain-id", rollappConfig.HubData.ID,
	}
}
