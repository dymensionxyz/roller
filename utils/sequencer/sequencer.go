package sequencer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/tx"
)

func Register(raCfg config.RollappConfig) error {
	seqPubKey, err := utils.GetSequencerPubKey(raCfg)
	if err != nil {
		return err
	}

	seqMetadataPath := filepath.Join(
		raCfg.Home,
		consts.ConfigDirName.Rollapp,
		"init",
		"sequencer-metadata.json",
	)
	_, err = isValidSequencerMetadata(seqMetadataPath)
	if err != nil {
		return err
	}

	seqMinBond, err := GetMinSequencerBond()
	if err != nil {
		return err
	}

	// TODO: handle raw_log
	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx",
		"sequencer",
		"create-sequencer",
		seqPubKey,
		raCfg.RollappID,
		seqMetadataPath,
		fmt.Sprintf("%s%s", seqMinBond.Amount.String(), seqMinBond.Denom),
		"--from", consts.KeysIds.HubSequencer,
		"--keyring-backend", "test",
		"--fees", fmt.Sprintf("%d%s", consts.DefaultFee, consts.Denoms.Hub),
		"--gas-adjustment", "1.3",
		"--keyring-dir", filepath.Join(utils.GetRollerRootDir(), consts.ConfigDirName.HubKeys),
	)

	txHash, err := bash.ExecCommandWithInput(cmd)
	if err != nil {
		return err
	}

	err = tx.MonitorTransaction(raCfg.HubData.RPC_URL, txHash)
	if err != nil {
		return err
	}

	return nil
}

func isValidSequencerMetadata(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}

	// nolint:errcheck
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return false, err
	}

	var sm dymensionseqtypes.SequencerMetadata
	err = json.Unmarshal(b, &sm)
	if err != nil {
		return false, err
	}

	return true, err
}

func GetMinSequencerBond() (*cosmossdktypes.Coin, error) {
	var qpr dymensionseqtypes.QueryParamsResponse
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "params", "-o", "json",
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	_ = json.Unmarshal(out.Bytes(), &qpr)

	return &qpr.Params.MinBond, nil
}

// TODO: dymd q sequencer show-sequencer could be used instead
func IsRegisteredAsSequencer(seq []Info, addr string) bool {
	if len(seq) == 0 {
		return false
	}

	return slices.ContainsFunc(
		seq,
		func(s Info) bool { return strings.Compare(s.Address, addr) == 0 },
	)
}

func GetLatestSnapshot(raID string, hd consts.HubData) (*SnapshotInfo, error) {
	sequencers, err := GetRegisteredSequencers(raID, hd)
	if err != nil {
		return nil, err
	}

	var latestSnapshot *SnapshotInfo
	maxHeight := 0

	for _, s := range sequencers.Sequencers {
		for _, snapshot := range s.Metadata.Snapshots {
			height, err := strconv.Atoi(snapshot.Height)
			if err != nil {
				continue
			}

			if height > maxHeight {
				maxHeight = height
				latestSnapshot = snapshot
			}
		}
	}

	return latestSnapshot, nil
}

func GetAllP2pPeers(raID string, hd consts.HubData) ([]string, error) {
	sequencers, err := GetRegisteredSequencers(raID, hd)
	if err != nil {
		return nil, err
	}

	var peers []string

	for _, s := range sequencers.Sequencers {
		if len(sequencers.Sequencers) > 1 {
			peers = append(peers, s.Metadata.P2PSeeds[0])
		} else {
			peers = append(peers, s.Metadata.P2PSeeds...)
		}
	}

	return peers, nil
}

func GetRegisteredSequencers(
	raID string, hd consts.HubData,
) (*Sequencers, error) {
	var seq Sequencers
	cmd := getShowSequencerByRollappCmd(raID, hd)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &seq)
	if err != nil {
		return nil, err
	}

	return &seq, nil
}

func GetMetadata(addr string, hd consts.HubData) (*Metadata, error) {
	var seqinfo Info

	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "show-sequencer", addr,
		"--node", hd.RPC_URL,
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &seqinfo)
	if err != nil {
		return nil, err
	}

	return &seqinfo.Metadata, nil
}

func getShowSequencerByRollappCmd(raID string, hd consts.HubData) *exec.Cmd {
	return exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "show-sequencers-by-rollapp",
		raID, "-o", "json", "--node", hd.RPC_URL,
	)
}
