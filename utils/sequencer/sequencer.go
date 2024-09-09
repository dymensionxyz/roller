package sequencer

import (
	"encoding/json"
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

func Register(raCfg config.RollappConfig, desiredBond string) error {
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

	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx",
		"sequencer",
		"create-sequencer",
		seqPubKey,
		raCfg.RollappID,
		desiredBond,
		seqMetadataPath,
		"--from", consts.KeysIds.HubSequencer,
		"--keyring-backend", "test",
		"--fees", "1dym",
		"--gas", "auto",
		"--gas-adjustment", "1.3",
		"--keyring-dir", filepath.Join(utils.GetRollerRootDir(), consts.ConfigDirName.HubKeys),
		"--node", raCfg.HubData.RPC_URL, "--chain-id", raCfg.HubData.ID,
	)

	txOutput, err := bash.ExecCommandWithInput(cmd, "signatures")
	if err != nil {
		return err
	}

	txHash, err := bash.ExtractTxHash(txOutput)
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

func GetSequencerAccountData(cfg config.RollappConfig) (string, error) {
	seqAddr, err := utils.GetAddressBinary(
		utils.KeyConfig{
			ID:  consts.KeysIds.HubSequencer,
			Dir: consts.ConfigDirName.HubKeys,
		}, cfg.Home,
	)
	if err != nil {
		return "", err
	}

	return seqAddr, nil
}

func GetMinSequencerBond(hd consts.HubData) (*cosmossdktypes.Coin, error) {
	var qpr dymensionseqtypes.QueryParamsResponse
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "params", "-o", "json", "--node", hd.RPC_URL, "--chain-id", hd.ID,
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
	sequencers, err := RegisteredRollappSequencersOnHub(raID, hd)
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
	sequencers, err := RegisteredRollappSequencersOnHub(raID, hd)
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

func RegisteredRollappSequencersOnHub(
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

func RegisteredRollappSequencers(
	raID string,
) (*Sequencers, error) {
	var seq Sequencers
	cmd := getShowSequencerCmd(raID)

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

func GetMetadata(
	addr string,
	hd consts.HubData,
) (*Metadata, error) {
	var seqinfo ShowSequencerResponse

	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "show-sequencer", addr,
		"--node", hd.RPC_URL, "-o", "json", "--chain-id", hd.ID,
	)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &seqinfo)
	if err != nil {
		return nil, err
	}

	return &seqinfo.Sequencer.Metadata, nil
}

func getShowSequencerByRollappCmd(raID string, hd consts.HubData) *exec.Cmd {
	return exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "show-sequencers-by-rollapp",
		raID, "-o", "json", "--node", hd.RPC_URL, "--chain-id", hd.ID,
	)
}

func getShowSequencerCmd(raID string) *exec.Cmd {
	return exec.Command(
		consts.Executables.RollappEVM,
		"q", "sequencers", "sequencers",
		"-o", "json", "--node", "http://localhost:26657", "--chain-id", raID,
	)
}

func GetHubSequencerAddress(cfg config.RollappConfig) (string, error) {
	seqAddr, err := utils.GetAddressBinary(
		utils.KeyConfig{
			ID:  consts.KeysIds.HubSequencer,
			Dir: consts.ConfigDirName.HubKeys,
		}, cfg.Home,
	)
	if err != nil {
		return "", err
	}

	return seqAddr, nil
}

func GetSequencerData(cfg config.RollappConfig) ([]utils.AccountData, error) {
	seqAddr, err := GetHubSequencerAddress(cfg)
	if err != nil {
		return nil, err
	}

	sequencerBalance, err := utils.QueryBalance(
		utils.ChainQueryConfig{
			Binary: consts.Executables.Dymension,
			Denom:  consts.Denoms.Hub,
			RPC:    cfg.HubData.RPC_URL,
		}, seqAddr,
	)
	if err != nil {
		return nil, err
	}
	return []utils.AccountData{
		{
			Address: seqAddr,
			Balance: sequencerBalance,
		},
	}, nil
}
