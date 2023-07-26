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
	output, err := utils.ExecBashCommandWithStdout(r.queryChannelsRollappCmd())
	if err != nil {
		return "", "", err
	}

	// While there are JSON objects in the stream...
	dec := json.NewDecoder(&output)
	for dec.More() {
		var outputStruct RollappQueryResult
		err = dec.Decode(&outputStruct)
		if err != nil {
			return "", "", fmt.Errorf("error while decoding JSON: %v", err)
		}

		if outputStruct.State == "STATE_OPEN" {
			output, err := utils.ExecBashCommandWithStdout(r.queryChannelsHubCmd(outputStruct.Counterparty.ChannelID))
			if err != nil {
				return "", "", err
			}

			var res HubQueryResult
			err = json.Unmarshal(output.Bytes(), &res)
			if err != nil {
				return "", "", err
			}

			if res.Channel.State != "STATE_OPEN" {
				fmt.Printf("channel %s is STATE_OPEN on the rollapp, but channel %s is %s on the hub\n",
					outputStruct.ChannelID, outputStruct.Counterparty.ChannelID, res.Channel.State,
				)
				continue
			}

			r.SrcChannel = outputStruct.ChannelID
			r.DstChannel = outputStruct.Counterparty.ChannelID
		}
	}

	return r.SrcChannel, r.DstChannel, nil
}

func (r *Relayer) queryChannelsRollappCmd() *exec.Cmd {
	args := []string{"q", "channels", r.RollappID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) queryChannelsHubCmd(channelID string) *exec.Cmd {
	args := []string{"q", "channel", r.HubID, channelID, "transfer"}
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
}

type ProofHeight struct {
	RevNumber string `json:"revision_number"`
	RevHeight string `json:"revision_height"`
}
type HubQueryResult struct {
	Channel     Output      `json:"channel"`
	Proof       string      `json:"proof"`
	ProofHeight ProofHeight `json:"proof_height"`
}

type RollappQueryResult struct {
	Output
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}
