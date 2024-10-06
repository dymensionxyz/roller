package rollapp

import (
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"

	dymensiontypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	bashutils "github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/version"
)

func GetCurrentHeight() (*BlockInformation, error) {
	cmd := getCurrentBlockCmd()
	out, err := bashutils.ExecCommandWithStdout(cmd)
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
	out, err := bashutils.ExecCommandWithStdout(cmd)
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
	_, err := bashutils.ExecCommandWithStdout(cmd)
	if err != nil {
		if strings.Contains(err.Error(), "NotFound") {
			return false, errors.New("rollapp not found ")
		}
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

type GetProposerResponse struct {
	ProposerAddr string `json:"proposerAddr"`
}

func GetCurrentProposerCmd(raID string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "proposer",
		raID, "-o", "json", "--node", hd.RPC_URL, "--chain-id", hd.ID,
	)

	return cmd
}

func GetCurrentProposer(raID string, hd consts.HubData) (string, error) {
	cmd := GetCurrentProposerCmd(raID, hd)

	out, err := bashutils.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}
	var resp GetProposerResponse

	err = json.Unmarshal(out.Bytes(), &resp)
	if err != nil {
		return "", err
	}

	return resp.ProposerAddr, nil
}

func RollappConfigDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config")
}

func GetRollappSequencerAddress(home string) (string, error) {
	seqKeyConfig := keys.KeyConfig{
		Dir:         consts.ConfigDirName.Rollapp,
		ID:          consts.KeysIds.RollappSequencer,
		ChainBinary: consts.Executables.RollappEVM,
		Type:        consts.EVM_ROLLAPP,
	}
	addr, err := keys.GetAddressBinary(seqKeyConfig, home)
	if err != nil {
		return "", err
	}

	return addr, nil
}

func GetRollappMetadataFromChain(
	home, raID string,
	hd *consts.HubData,
) (*roller.RollappConfig, error) {
	var cfg roller.RollappConfig
	var raResponse ShowRollappResponse

	getRollappCmd := GetRollappCmd(raID, *hd)

	out, err := bash.ExecCommandWithStdout(getRollappCmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &raResponse)
	if err != nil {
		return nil, err
	}

	vmt, _ := consts.ToVMType(strings.ToLower(raResponse.Rollapp.VmType))

	var DA consts.DaData

	switch hd.ID {
	case consts.PlaygroundHubID:
		DA = consts.DaNetworks[string(consts.CelestiaTestnet)]
	// case consts.MainnetHubID:
	// 	DA = consts.DaNetworks[string(consts.CelestiaMainnet)]
	default:
		fmt.Println("unsupported Hub: ", hd.ID)
	}

	cfg = roller.RollappConfig{
		Home:                 home,
		GenesisHash:          raResponse.Rollapp.GenesisInfo.GenesisChecksum,
		GenesisUrl:           raResponse.Rollapp.Metadata.GenesisUrl,
		RollappID:            raResponse.Rollapp.RollappId,
		RollappBinary:        consts.Executables.RollappEVM,
		RollappVMType:        vmt,
		Denom:                raResponse.Rollapp.GenesisInfo.NativeDenom.Base,
		Decimals:             18,
		HubData:              *hd,
		DA:                   DA,
		RollerVersion:        "latest",
		Environment:          hd.ID,
		RollappBinaryVersion: version.BuildVersion,
		Bech32Prefix:         raResponse.Rollapp.GenesisInfo.Bech32Prefix,
		BaseDenom:            "",
		MinGasPrices:         "0",
	}

	return &cfg, nil
}
