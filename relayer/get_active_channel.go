package relayer

import (
	"encoding/json"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"os/exec"
)

func GetConnectionChannels(dstConnectionID string, rollappConfig config.RollappConfig) (
	ConnectionChannels, error) {
	commonDymdFlags := utils.GetCommonDymdFlags(rollappConfig)
	args := []string{"query", "ibc", "channel", "connections", dstConnectionID}
	args = append(args, commonDymdFlags...)
	cmd := exec.Command(consts.Executables.Dymension, args...)
	out, err := cmd.Output()
	if err != nil {
		return ConnectionChannels{}, err
	}
	channels, err := extractChannelsFromResponse(out)
	if err != nil {
		return ConnectionChannels{}, err
	}
	return channels, nil
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

func extractChannelsFromResponse(jsonData []byte) (ConnectionChannels, error) {
	var channels ChannelList
	if err := json.Unmarshal(jsonData, &channels); err != nil {
		return ConnectionChannels{}, err
	}
	for _, channel := range channels.Channels {
		if channel.State == "STATE_OPEN" {
			return ConnectionChannels{
				Src: channel.Counterparty.ChannelID,
				Dst: channel.ChannelID,
			}, nil
		}
	}
	return ConnectionChannels{}, nil
}
