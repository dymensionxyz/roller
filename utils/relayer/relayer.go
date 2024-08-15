package relayer

import (
	"encoding/json"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

type Channels struct {
	Channels []struct {
		State        string `json:"state"`
		Ordering     string `json:"ordering"`
		Counterparty struct {
			PortId    string `json:"port_id"`
			ChannelId string `json:"channel_id"`
		} `json:"counterparty"`
		ConnectionHops []string `json:"connection_hops"`
		Version        string   `json:"version"`
		PortId         string   `json:"port_id"`
		ChannelId      string   `json:"channel_id"`
	} `json:"channels"`
	Pagination struct {
		NextKey interface{} `json:"next_key"`
		Total   string      `json:"total"`
	} `json:"pagination"`
	Height struct {
		RevisionNumber string `json:"revision_number"`
		RevisionHeight string `json:"revision_height"`
	} `json:"height"`
}

func GetRegisteredSequencers(
	raID string, hd consts.HubData,
) (*Channels, error) {
	var ibcChannels Channels
	cmd := GetQueryRollappIBCChannels()

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &ibcChannels)
	if err != nil {
		return nil, err
	}

	return &ibcChannels, nil
}

func GetQueryRollappIBCChannels() *exec.Cmd {
	return exec.Command(
		consts.Executables.RollappEVM,
		"q", "ibc", "channels",
		"-o", "json",
	)
}
