package start

import (
	"encoding/json"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"os/exec"
)

// GetSourceChannelForConnection Returns the open source channel for the given destination connection ID. If no open channel exists, it returns an
// empty string.
func GetSourceChannelForConnection(dstConnectionID string, rollappConfig utils.RollappConfig) (string, error) {
	commonDymdFlags := utils.GetCommonDymdFlags(rollappConfig)
	args := []string{"query", "ibc", "channel", "connections", dstConnectionID}
	args = append(args, commonDymdFlags...)
	cmd := exec.Command(consts.Executables.Dymension, args...)
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	channelId, err := GetOpenStateChannelID(out)
	if err != nil {
		return "", err
	}
	return channelId, nil
}

type Channel struct {
	State        string `json:"state"`
	Counterparty struct {
		ChannelID string `json:"channel_id"`
	} `json:"counterparty"`
}

type ChannelList struct {
	Channels []Channel `json:"channels"`
}

func GetOpenStateChannelID(jsonData []byte) (string, error) {
	var channels ChannelList
	if err := json.Unmarshal(jsonData, &channels); err != nil {
		return "", err
	}

	for _, channel := range channels.Channels {
		if channel.State == "STATE_OPEN" {
			return channel.Counterparty.ChannelID, nil
		}
	}
	return "", nil
}
