package sequencer

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"sync"
)

type Sequencer struct {
	RlpCfg      config.RollappConfig
	RPCPort     string
	APIPort     string
	JsonRPCPort string
	logger      *log.Logger
}

var instance *Sequencer
var once sync.Once

func GetInstance(rlpCfg config.RollappConfig) *Sequencer {
	once.Do(func() {
		seq := &Sequencer{
			logger: log.New(io.Discard, "", 0),
			RlpCfg: rlpCfg,
		}
		if err := seq.ReadPorts(); err != nil {
			panic(err)
		}
		instance = seq
	})
	return instance
}

func (seq *Sequencer) GetStartCmd() *exec.Cmd {
	rollappConfigDir := filepath.Join(seq.RlpCfg.Home, consts.ConfigDirName.Rollapp)
	cmd := exec.Command(
		seq.RlpCfg.RollappBinary,
		"start",
		"--home", rollappConfigDir,
		"--log-file", filepath.Join(rollappConfigDir, "rollapp.log"),
	)
	return cmd
}
