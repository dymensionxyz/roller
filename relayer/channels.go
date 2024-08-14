package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
)

// TODO: Change to use the connection for fetching relevant channel using connection-channels rly command
func (r *Relayer) LoadActiveChannel() (string, string, error) {
	fmt.Println("inside")
	var outputStruct RollappQueryResult
	var foundOpenChannel RollappQueryResult

	var activeConnectionID string
	activeConnectionID, err := r.GetActiveConnection()
	if err != nil {
		if keyErr, ok := err.(*utils.KeyNotFoundError); ok {
			r.logger.Printf("No active connection found. Key not found: %v", keyErr)
			return "", "", nil
		} else {
			r.logger.Println("another err", err)
			return "", "", err
		}
	}
	if activeConnectionID == "" {
		return "", "", nil
	}
	fmt.Println(activeConnectionID)

	output, err := bash.ExecCommandWithStdout(r.queryChannelsRollappCmd(activeConnectionID))
	if err != nil {
		return "", "", err
	}

	fmt.Println(output)

	if output.Len() == 0 {
		return "", "", nil
	}

	dec := json.NewDecoder(&output)
	for dec.More() {
		err = dec.Decode(&outputStruct)
		if err != nil {
			return "", "", fmt.Errorf("error while decoding JSON: %v", err)
		}

		if outputStruct.ConnectionHops[0] != activeConnectionID {
			r.logger.Printf(
				"skipping channel %s as it's not on the active connection %s",
				outputStruct.ChannelID, activeConnectionID,
			)
			continue
		}

		if outputStruct.State != "STATE_OPEN" {
			continue
		}

		j, _ := json.MarshalIndent(outputStruct, "", "  ")
		fmt.Println(string(j))

		// found STATE_OPEN channel
		// Check if the channel is open on the hub
		var res HubQueryResult
		outputHub, err := bash.ExecCommandWithStdout(
			r.queryChannelsHubCmd(outputStruct.ChannelID),
		)
		if err != nil {
			return "", "", err
		}

		err = json.Unmarshal(outputHub.Bytes(), &res)
		if err != nil {
			return "", "", err
		}

		if res.Channel.State != "STATE_OPEN" {
			r.logger.Printf(
				"channel %s is STATE_OPEN on the rollapp, but channel %s is %s on the hub",
				outputStruct.ChannelID,
				outputStruct.Counterparty.ChannelID,
				res.Channel.State,
			)
			continue
		}

		// Found open channel on both ends
		foundOpenChannel = outputStruct
		fmt.Println("found", foundOpenChannel)
		break
	}

	fmt.Println(foundOpenChannel)

	r.SrcChannel = foundOpenChannel.ChannelID
	r.DstChannel = foundOpenChannel.Counterparty.ChannelID
	return "", "", nil
}

func (r *Relayer) queryChannelsRollappCmd(connectionID string) *exec.Cmd {
	args := []string{"q", "connection-channels", r.RollappID, connectionID}
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
	PortID       string `json:"port_id"`
	ChannelID    string `json:"channel_id"`
	ChainID      string `json:"chain_id"`
	ClientID     string `json:"client_id"`
	ConnectionID string `json:"connection_id"`
}

type Output struct {
	State          string       `json:"state"`
	Ordering       string       `json:"ordering"`
	Counterparty   Counterparty `json:"counterparty"`
	ConnectionHops []string     `json:"connection_hops"`
	Version        string       `json:"version"`
	ChainID        string       `json:"chain_id"`
	ChannelID      string       `json:"channel_id"`
	ClientID       string       `json:"client_id"`
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
