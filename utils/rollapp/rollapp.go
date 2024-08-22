package rollapp

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	dymensiontypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/roller/cmd/consts"
	globalutils "github.com/dymensionxyz/roller/utils/bash"
)

func GetCurrentHeight() (*BlockInformation, error) {
	cmd := getCurrentBlockCmd()
	out, err := globalutils.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	var blockInfo BlockInformation
	err = json.Unmarshal(out.Bytes(), &blockInfo)
	if err != nil {
		return nil, err
	}

	return &blockInfo, nil
}

func getCurrentBlockCmd() *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"q",
		"block",
	)
	return cmd
}

func GetInitialSequencerAddress(raID string, hd consts.HubData) (string, error) {
	cmd := GetShowRollappCmd(raID, hd)
	out, err := globalutils.ExecCommandWithStdout(cmd)
	if err != nil {
		fmt.Println(err)
	}

	var ra dymensiontypes.QueryGetRollappResponse
	_ = json.Unmarshal(out.Bytes(), &ra)

	return ra.Rollapp.InitialSequencer, nil
}

func IsInitialSequencer(addr, raID string, hd consts.HubData) (bool, error) {
	initSeqAddr, err := GetInitialSequencerAddress(raID, hd)
	if err != nil {
		return false, err
	}

	fmt.Printf("%s\n%s\n", addr, initSeqAddr)

	if strings.TrimSpace(addr) == strings.TrimSpace(initSeqAddr) {
		return true, nil
	}

	return false, nil
}

// TODO: most of rollapp utility functions should be tied to an entity
func IsRollappRegistered(raID string, hd consts.HubData) (bool, error) {
	cmd := GetShowRollappCmd(raID, hd)
	_, err := globalutils.ExecCommandWithStdout(cmd)
	if err != nil {
		// tODO: handle NotFound error
		return false, err
	}

	return true, nil
}

func GetShowRollappCmd(raID string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"rollapp",
		"show",
		raID,
		"-o", "json",
		"--node", hd.RPC_URL,
		"--chain-id", hd.ID,
	)

	return cmd
}

func GetRollappCmd(raID string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "rollapp", "show",
		raID, "-o", "json", "--node", hd.RPC_URL, "--chain-id", hd.ID,
	)

	return cmd
}

func RollappConfigDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config")
}
