package sequencer

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
)

// TODO(FIXME): Assumptions on rollapp price and keyring backend
func (seq *Sequencer) GetSendCmd(destAddress string) *exec.Cmd {
	rollappConfigDir := filepath.Join(seq.RlpCfg.Home, consts.ConfigDirName.Rollapp)
	cmd := exec.Command(
		seq.RlpCfg.RollappBinary,
		"tx", "bank", "send",
		consts.KeysIds.RollappSequencer, destAddress, "1"+seq.RlpCfg.Denom,
		"--home", rollappConfigDir,
		"--broadcast-mode", "sync",
		"--keyring-backend", string(seq.RlpCfg.KeyringBackend),
		"--yes",
	)
	return cmd
}
