package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func (r *Relayer) LoadChannels() (string, string, error) {
	output, err := utils.ExecBashCommand(r.queryChannelsCmd())
	if err != nil {
		return "", "", err
	}

	dec := json.NewDecoder(&output)

	// While there are JSON objects in the stream...
	var stateOpenChannel Output

	for dec.More() {
		var outputStruct Output
		err = dec.Decode(&outputStruct)
		if err != nil {
			return "", "", fmt.Errorf("error while decoding JSON: %v", err)
		}

		if outputStruct.State == "STATE_OPEN" {
			//we want the last open channel
			stateOpenChannel = outputStruct
		}
	}

	//TODO validate with the connection ID with the path

	r.SrcChannel = stateOpenChannel.ChannelID
	r.DstChannel = stateOpenChannel.Counterparty.ChannelID
	return r.SrcChannel, r.DstChannel, nil
}

func (r *Relayer) queryChannelsCmd() *exec.Cmd {
	args := []string{"q", "channels", r.RollappID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) ChannelReady() bool {
	return r.SrcChannel != "" && r.DstChannel != ""
}

type Counterparty struct {
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}

type Output struct {
	State          string       `json:"state"`
	Ordering       string       `json:"ordering"`
	Counterparty   Counterparty `json:"counterparty"`
	ConnectionHops []string     `json:"connection_hops"`
	Version        string       `json:"version"`
	PortID         string       `json:"port_id"`
	ChannelID      string       `json:"channel_id"`
}
