package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	cmdutils "github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils"
)

// TODO: Change to use the connection for fetching relevant channel using connection-channels rly command
func (r *Relayer) LoadActiveChannel() (string, string, error) {
	output, err := cmdutils.ExecBashCommandWithStdout(r.queryChannelsRollappCmd())
	if err != nil {
		return "", "", err
	}

	if output.Len() == 0 {
		return "", "", nil
	}

	// While there are JSON objects in the stream...
	var outputStruct RollappQueryResult
	var foundOpenChannel RollappQueryResult
	var activeConnectionID string

	activeConnectionID, err = r.GetActiveConnection()
	if err != nil {
		if keyErr, ok := err.(*utils.KeyNotFoundError); ok {
			r.logger.Printf("No active connection found. Key not found: %v", keyErr)
			return "", "", nil
		} else {
			return "", "", err
		}
	}
	if activeConnectionID == "" {
		return "", "", nil
	}

	dec := json.NewDecoder(&output)
	for dec.More() {
		err = dec.Decode(&outputStruct)
		if err != nil {
			return "", "", fmt.Errorf("error while decoding JSON: %v", err)
		}
		if outputStruct.ConnectionHops[0] != activeConnectionID {
			r.logger.Printf("skipping channel %s as it's not on the active connection %s",
				outputStruct.ChannelID, activeConnectionID)
			continue
		}

		if outputStruct.State != "STATE_OPEN" {
			continue
		}

		// found STATE_OPEN channel
		// Check if the channel is open on the hub
		var res HubQueryResult
		outputHub, err := cmdutils.ExecBashCommandWithStdout(r.queryChannelsHubCmd(outputStruct.Counterparty.ChannelID))
		if err != nil {
			return "", "", err
		}
		err = json.Unmarshal(outputHub.Bytes(), &res)
		if err != nil {
			return "", "", err
		}

		if res.Channel.State != "STATE_OPEN" {
			r.logger.Printf("channel %s is STATE_OPEN on the rollapp, but channel %s is %s on the hub",
				outputStruct.ChannelID, outputStruct.Counterparty.ChannelID, res.Channel.State,
			)
			continue
		}

		// Found open channel on both ends
		foundOpenChannel = outputStruct
		break
	}

	r.SrcChannel = foundOpenChannel.ChannelID
	r.DstChannel = foundOpenChannel.Counterparty.ChannelID
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
