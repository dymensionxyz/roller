package sequencer

import (
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"sync"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
)

type Sequencer struct {
	RlpCfg      roller.RollappConfig
	RPCPort     string
	APIPort     string
	JsonRPCPort string
	logger      *log.Logger
}

var (
	instance *Sequencer
	once     sync.Once
)

func GetInstance(rlpCfg roller.RollappConfig) *Sequencer {
	once.Do(
		func() {
			seq := &Sequencer{
				logger: log.New(io.Discard, "", 0),
				RlpCfg: rlpCfg,
			}
			if err := seq.ReadPorts(); err != nil {
				panic(err)
			}
			instance = seq
		},
	)
	return instance
}

func (seq *Sequencer) GetStartCmd(logLevel string, keyringBackend consts.SupportedKeyringBackend) *exec.Cmd {
	rollappConfigDir := filepath.Join(seq.RlpCfg.Home, consts.ConfigDirName.Rollapp)
	args := []string{
		"start",
		"--home", rollappConfigDir,
	}

	debugArgs := []string{"--log_level", logLevel}
	args = append(args, debugArgs...)

	cmd := exec.Command(
		consts.Executables.RollappEVM, args...,
	)
	return cmd
}
