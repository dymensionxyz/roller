package rollapp

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"

	dymensiontypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	globalutils "github.com/dymensionxyz/roller/utils/bash"
	tmtypes "github.com/tendermint/tendermint/types"

	"github.com/dymensionxyz/roller/cmd/consts"
	sequencerutils "github.com/dymensionxyz/roller/utils/sequencer"
)

func GetCurrentHeight() (*BlockInformation, error) {
	cmd := getCurrentBlockCmd()
	out, err := globalutils.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, nil
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

func GetInitialSequencerAddress(raID string) (string, error) {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"rollapp",
		"show",
		raID,
		"-o",
		"json",
	)

	out, err := globalutils.ExecCommandWithStdout(cmd)
	if err != nil {
		fmt.Println(err)
	}

	var ra dymensiontypes.QueryGetRollappResponse
	_ = json.Unmarshal(out.Bytes(), &ra)

	return ra.Rollapp.InitialSequencer, nil
}

func IsInitialSequencer(addr, raID string) (bool, error) {
	initSeqAddr, err := GetInitialSequencerAddress(raID)
	if err != nil {
		return false, err
	}

	fmt.Printf("%s\n%s\n", addr, initSeqAddr)

	if strings.TrimSpace(addr) == strings.TrimSpace(initSeqAddr) {
		return true, nil
	}

	return false, nil
}

func GetRegisteredSequencers(
	raID string,
) (*sequencerutils.Sequencers, error) {
	var seq sequencerutils.Sequencers
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"sequencer",
		"show-sequencers-by-rollapp",
		raID,
		"--output", "json",
	)

	out, err := globalutils.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &seq)
	if err != nil {
		return nil, err
	}

	return &seq, nil
}

type BlockInformation struct {
	BlockId tmtypes.BlockID `json:"block_id"`
	Block   tmtypes.Block   `json:"block"`
}
