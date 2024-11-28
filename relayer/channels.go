package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"slices"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	dymintutils "github.com/dymensionxyz/roller/utils/dymint"
	"github.com/dymensionxyz/roller/utils/roller"
)

// TODO: Change to use the connection for fetching relevant channel using connection-channels rly command
func (r *Relayer) LoadActiveChannel(
	raData consts.RollappData,
	hd consts.HubData,
) error {
	spinner, _ := pterm.DefaultSpinner.Start("loading active IBC channels")
	defer spinner.Stop()

	gacCmd := r.getQueryChannelsRollappCmd(raData)

	gacOut, err := bash.ExecCommandWithStdout(gacCmd)
	if err != nil {
		return err
	}

	var gacResponse QueryChannelsResponse
	err = json.Unmarshal(gacOut.Bytes(), &gacResponse)
	if err != nil {
		return err
	}

	if len(gacResponse.Channels) == 0 {
		return ErrNoOpenChannel
	}

	raIbcChanIndex := slices.IndexFunc(
		gacResponse.Channels, func(ibcChan Channel) bool {
			return ibcChan.State == "STATE_OPEN"
		},
	)

	fmt.Println(raIbcChanIndex)

	if raIbcChanIndex == -1 {
		return ErrNoOpenChannel
	}

	raChan := gacResponse.Channels[raIbcChanIndex]
	r.SrcChannel = raChan.Counterparty.ChannelID
	r.DstChannel = raChan.ChannelID
	r.DstConnectionID = raChan.ConnectionHops[0]

	return nil
}

// func (r *Relayer) GetActiveConnections(
// 	raData consts.RollappData,
// 	hd consts.HubData,
// ) (string, string, error) {
// 	var gacResponse QueryChannelsResponse
//
// 	pterm.Info.Println("querying connections")
// 	cmd := r.queryConnectionRollappCmd(raData)
// 	rollappConnectionOutput, err := bash.ExecCommandWithStdout(cmd)
// 	if err != nil {
// 		pterm.Error.Printfln(
// 			"failed to find connection on the rollapp side for %s: %v",
// 			r.RollappID,
// 			err,
// 		)
// 		return "", "", err
// 	}
// 	var raIbcConnections ConnectionsQueryResult
// 	err = json.Unmarshal(rollappConnectionOutput.Bytes(), &raIbcConnections)
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	activeIbcConnectionID := gacResponse.Channels[0].ConnectionHops[0]
// 	raIbcConnInx := slices.IndexFunc(
// 		raIbcConnections.Connections, func(conn ConnectionInfo) bool {
// 			return conn.ID == activeIbcConnectionID
// 		},
// 	)
//
// 	raIbcConn := raIbcConnections.Connections[raIbcConnInx]
// 	j, _ := json.Marshal(raIbcConn)
// 	fmt.Printf("\tRA IBC Connection:\n%s", string(j))
// 	// END QUERY CONNECTIONS
//
// 	return "", "", errors.New("debugging")
//
// 	var activeRaConnectionID string
// 	var activeHubConnectionID string
// 	activeRaConnectionID, activeHubConnectionID, err = r.GetActiveConnectionIDs(raData, hd)
// 	if err != nil {
// 		if keyErr, ok := err.(*utils.KeyNotFoundError); ok {
// 			r.logger.Printf("No active connection found. Key not found: %v", keyErr)
// 			return "", "", nil
// 		} else {
// 			r.logger.Println("something bad happened", err)
// 			return "", "", err
// 		}
// 	}
// 	if activeRaConnectionID == "" {
// 		r.logger.Println("no active connection found")
// 		return "", "", nil
// 	}
//
// 	pterm.Info.Println("active connection found on the hub side: ", activeHubConnectionID)
// 	pterm.Info.Println("active connection found on the rollapp side: ", activeRaConnectionID)
//
// 	var raChannelResponse QueryChannelsResponse
// 	rollappChannels, err := bash.ExecCommandWithStdout(r.queryChannelsRollappCmd(raData))
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	err = json.Unmarshal(rollappChannels.Bytes(), &raChannelResponse)
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	if len(raChannelResponse.Channels) == 0 {
// 		return "", "", nil
// 	}
//
// 	for _, v := range raChannelResponse.Channels {
// 		fmt.Printf("%s: %s\n", v.ChannelID, v.State)
// 	}
//
// 	raChanIndex := slices.IndexFunc(
// 		raChannelResponse.Channels, func(ibcChan Channel) bool {
// 			return ibcChan.ConnectionHops[0] == activeRaConnectionID &&
// 				ibcChan.State == "STATE_OPEN"
// 		},
// 	)
// 	raChan := raChannelResponse.Channels[raChanIndex]
//
// 	var hubChannelResponse QueryChannelsResponse
// 	hubChannels, err := bash.ExecCommandWithStdout(r.queryChannelsHubCmd(hd))
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	err = json.Unmarshal(hubChannels.Bytes(), &hubChannelResponse)
// 	if err != nil {
// 		return "", "", err
// 	}
//
// 	if len(hubChannelResponse.Channels) == 0 {
// 		return "", "", nil
// 	}
//
// 	hubChanIndex := slices.IndexFunc(
// 		hubChannelResponse.Channels, func(ibcChan Channel) bool {
// 			return ibcChan.ConnectionHops[0] == activeHubConnectionID &&
// 				ibcChan.State == "STATE_OPEN"
// 		},
// 	)
// 	hubChan := hubChannelResponse.Channels[hubChanIndex]
//
// 	pterm.Info.Println("active channel found on the hub side: ", hubChan.ChannelID)
// 	pterm.Info.Println("active channel found on the rollapp side: ", raChan.ChannelID)
//
// 	spinner.Success("IBC channels loaded successfully")
//
// 	r.SrcChannel = hubChan.ChannelID
// 	r.DstChannel = raChan.ChannelID
//
// 	return r.SrcChannel, r.DstChannel, nil
// }

func (r *Relayer) getQueryChannelsRollappCmd(raData consts.RollappData) *exec.Cmd {
	args := []string{"q", "ibc", "channel", "channels"}
	args = append(args, "--node", raData.RpcUrl, "--chain-id", raData.ID, "-o", "json")

	cmd := exec.Command(consts.Executables.RollappEVM, args...)

	return cmd
}

func (r *Relayer) queryChannelsHubCmd(hd consts.HubData) *exec.Cmd {
	args := []string{"q", "ibc", "channel", "channels"}
	args = append(
		args,
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
		"-o",
		"json",
		"--limit",
		"100000",
	)

	cmd := exec.Command(consts.Executables.Dymension, args...)

	return cmd
}

func (r *Relayer) ChannelReady() bool {
	return r.SrcChannel != "" && r.DstChannel != ""
}

type QueryChannelsResponse struct {
	Channels   []Channel  `json:"channels"`
	Pagination Pagination `json:"pagination"`
	Height     Height     `json:"height"`
}

type Channel struct {
	State          string       `json:"state"`
	Ordering       string       `json:"ordering"`
	Counterparty   Counterparty `json:"counterparty"`
	ConnectionHops []string     `json:"connection_hops"`
	Version        string       `json:"version"`
	PortID         string       `json:"port_id"`
	ChannelID      string       `json:"channel_id"`
}

type Counterparty struct {
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}

type Pagination struct {
	NextKey interface{} `json:"next_key"`
	Total   string      `json:"total"`
}

type Height struct {
	RevisionNumber string `json:"revision_number"`
	RevisionHeight string `json:"revision_height"`
}

// @20241022
// The code below represents responses from the 'rly' command
// as of v1.16.* of roller, the information is fetched from the chain
// using the `rollappd` and `dymd` binaries respectfully.
// this code is left here in case it's required to perform actions
// specific to the 'rly' binary
type RlyCounterparty struct {
	PortID       string `json:"port_id"`
	ChannelID    string `json:"channel_id"`
	ChainID      string `json:"chain_id"`
	ClientID     string `json:"client_id"`
	ConnectionID string `json:"connection_id"`
}

type RlyOutput struct {
	State          string          `json:"state"`
	Ordering       string          `json:"ordering"`
	Counterparty   RlyCounterparty `json:"counterparty"`
	ConnectionHops []string        `json:"connection_hops"`
	Version        string          `json:"version"`
	ChainID        string          `json:"chain_id"`
	ChannelID      string          `json:"channel_id"`
	ClientID       string          `json:"client_id"`
}

type RlyProofHeight struct {
	RevNumber string `json:"revision_number"`
	RevHeight string `json:"revision_height"`
}

type RlyHubQueryResult struct {
	Channel     RlyOutput      `json:"channel"`
	Proof       string         `json:"proof"`
	ProofHeight RlyProofHeight `json:"proof_height"`
}

type RlyRollappQueryResult struct {
	RlyOutput
	PortID    string `json:"port_id"`
	ChannelID string `json:"channel_id"`
}

func (r *Relayer) HandleIbcChannelCreation(
	home string,
	rollappChainData roller.RollappConfig,
	logFileOption bash.CommandOption,
) error {
	defer func() {
		pterm.Info.Println("reverting dymint config to 1h")
		err := dymintutils.UpdateDymintConfigForIBC(home, "1h0m0s", true)
		if err != nil {
			pterm.Error.Println("failed to update dymint config: ", err)
			return
		}
	}()

	seq := sequencer.GetInstance(rollappChainData)

	pterm.Info.Println("setting block time to 5s for esstablishing IBC connection")
	err := dymintutils.UpdateDymintConfigForIBC(home, "5s", true)
	if err != nil {
		pterm.Error.Println("failed to update dymint config: ", err)
		return err
	}

	dymintutils.WaitForHealthyRollApp("http://localhost:26657/health")
	err = WaitForValidRollappHeight(seq)
	if err != nil {
		pterm.Error.Printf("rollapp did not reach valid height: %v\n", err)
		return err
	}

	err = VerifyRelayerBalances(r.Hub)
	if err != nil {
		pterm.Error.Printf("failed to verify relayer balances: %v\n", err)
		return err
	}

	pterm.Info.Println("establishing IBC transfer channel")
	channels, err := r.CreateIBCChannel(
		logFileOption,
		r.Rollapp,
		r.Hub,
	)
	if err != nil {
		pterm.Error.Printf("failed to create IBC channel: %v\n", err)
		return err
	}

	r.SrcChannel = channels.Src
	r.DstChannel = channels.Dst

	status := fmt.Sprintf(
		"Active\nrollapp: %s\n<->\nhub: %s",
		r.SrcChannel,
		r.DstChannel,
	)

	pterm.Info.Println(status)
	err = r.WriteRelayerStatus(status)
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}
