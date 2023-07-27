package relayer

import (
	"fmt"
	"log"

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
	}
}

func (r *Relayer) SetLogger(logger *log.Logger) {
	r.logger = logger
}

func (r *Relayer) GetRelayerStatus(config.RollappConfig) string {
	if r.ChannelReady() {
		return fmt.Sprintf("Active src, %s <-> %s, dst", r.SrcChannel, r.DstChannel)
	}

	_, _, _ = r.LoadChannels()
	return fmt.Sprintf("Starting...")
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
