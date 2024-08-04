package sequencer

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	dymensionseqtypes "github.com/dymensionxyz/dymension/v3/x/sequencer/types"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
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

	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx",
		"sequencer",
		"create-sequencer",
		seqPubKey,
		raCfg.RollappID,
		seqMetadataPath,
		seqMinBond,
		"--from", consts.KeysIds.HubSequencer,
		"--keyring-backend", "test",
		"--fees", "1000000000000000000adym",
		"--keyring-dir", filepath.Join(utils.GetRollerRootDir(), consts.ConfigDirName.HubKeys),
		"-y",
	)

	out, err := utils.ExecBashCommandWithStdout(cmd)
	if err != nil {
		return err
	}

	fmt.Println(out.String())
	return nil
}

func isValidSequencerMetadata(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}

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

func GetMinSequencerBond() (string, error) {
	var qpr dymensionseqtypes.QueryParamsResponse
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q", "sequencer", "params", "-o", "json",
	)

	out, err := utils.ExecBashCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}

	_ = json.Unmarshal(out.Bytes(), &qpr)

	minBond := fmt.Sprintf("%s%s", qpr.Params.MinBond.Amount, qpr.Params.MinBond.Denom)
	return minBond, nil
}

// TODO: dymd q sequencer show-sequencer could be used instead
func IsRegisteredAsSequencer(seq []Info, addr string) bool {
	return slices.ContainsFunc(
		seq,
		func(s Info) bool { return strings.Compare(s.Address, addr) == 0 },
	)
}
