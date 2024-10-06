package relayer

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
)

type Relayer struct {
	Home       string
	RollappID  string
	HubID      string
	SrcChannel string
	DstChannel string
	logger     *log.Logger
}

func NewRelayer(home, rollappID, hubID string) *Relayer {
	return &Relayer{
		Home:      home,
		RollappID: rollappID,
		HubID:     hubID,
		logger:    log.New(io.Discard, "", 0),
	}
}

func (r *Relayer) SetLogger(logger *log.Logger) {
	r.logger = logger
}

// TODO: review the servicemanager.Service implementation
func (r *Relayer) GetRelayerStatus(roller.RollappConfig) string {
	if r.ChannelReady() {
		return fmt.Sprintf(
			"Active Channels:\nrollapp: %s\n<->\nhub: %s",
			r.SrcChannel,
			r.DstChannel,
		)
	}
	bytes, err := os.ReadFile(r.StatusFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "Starting..."
		}
	}
	return string(bytes)
}

func (r *Relayer) WriteRelayerStatus(status string) error {
	// nolint:gofumpt
	return os.WriteFile(r.StatusFilePath(), []byte(status), 0o644)
}

func (r *Relayer) StatusFilePath() string {
	return filepath.Join(r.Home, consts.ConfigDirName.Relayer, "relayer_status.txt")
}

type ConnectionChannels struct {
	Src string
	Dst string
}
