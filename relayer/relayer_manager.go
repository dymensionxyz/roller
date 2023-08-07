package relayer

import (
	"errors"
	"github.com/dymensionxyz/roller/cmd/consts"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/config"
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

func (r *Relayer) GetRelayerStatus(config.RollappConfig) string {
	bytes, err := os.ReadFile(r.statusFilePath())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "Starting..."
		}
	}
	return string(bytes)
}

func (r *Relayer) WriteRelayerStatus(status string) error {
	return os.WriteFile(r.statusFilePath(), []byte(status), 0644)
}

func (r *Relayer) statusFilePath() string {
	return filepath.Join(r.Home, consts.ConfigDirName.Relayer, "relayer_status.txt")
}

type Channel struct {
	State        string `json:"state"`
	ChannelID    string `json:"channel_id"`
	Counterparty struct {
		ChannelID string `json:"channel_id"`
	} `json:"counterparty"`
}

type ChannelList struct {
	Channels []Channel `json:"channels"`
}

type ConnectionChannels struct {
	Src string
	Dst string
}
