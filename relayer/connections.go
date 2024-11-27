package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

type ConnectionsQueryResult struct {
	Connections []ConnectionInfo `json:"connections"`
	Height      ProofHeightInfo  `json:"height"`
	Pagination  PaginationInfo   `json:"pagination"`
}

type RlyConnectionsQueryResult struct {
	ID           string           `json:"id"`
	ClientID     string           `json:"client_id"`
	Versions     []VersionInfo    `json:"versions"`
	State        string           `json:"state"`
	Counterparty CounterpartyInfo `json:"counterparty"`
	DelayPeriod  string           `json:"delay_period"`
}

type RlyConnectionQueryResult struct {
	Connection  ConnectionInfo  `json:"connection"`
	Proof       string          `json:"proof"`
	ProofHeight ProofHeightInfo `json:"proof_height"`
}

type ConnectionInfo struct {
	ClientID     string           `json:"client_id"`
	Versions     []VersionInfo    `json:"versions"`
	State        string           `json:"state"`
	ID           string           `json:"id"`
	Counterparty CounterpartyInfo `json:"counterparty"`
	DelayPeriod  string           `json:"delay_period"`
}

type ProofHeightInfo struct {
	RevisionNumber string `json:"revision_number"`
	RevisionHeight string `json:"revision_height"`
}

type PaginationInfo struct {
	NextKey string `json:"next_key"`
	Total   string `json:"total"`
}

type VersionInfo struct {
	Identifier string   `json:"identifier"`
	Features   []string `json:"features"`
}

type CounterpartyInfo struct {
	ClientID     string     `json:"client_id"`
	ConnectionID string     `json:"connection_id"`
	Prefix       PrefixInfo `json:"prefix"`
}

type PrefixInfo struct {
	KeyPrefix string `json:"key_prefix"`
}

func (r *Relayer) HubIbcConnectionFromRaConnID(
	hd consts.HubData,
	raIbcConnectionID string,
) error {
	hubIbcConnections, err := r.HubIbcConnections(hd)
	if err != nil {
		return err
	}

	if len(hubIbcConnections.Connections) == 0 {
		return fmt.Errorf("no connections found on the rollapp side for %s", r.Rollapp.ID)
	}

	hubIbcConnIndex := slices.IndexFunc(
		hubIbcConnections.Connections, func(ibcConn ConnectionInfo) bool {
			return ibcConn.Counterparty.ConnectionID == raIbcConnectionID
		},
	)

	if hubIbcConnIndex == -1 {
		return fmt.Errorf("no open channel found for %s", r.Rollapp.ID)
	}

	j, _ := json.MarshalIndent(hubIbcConnections.Connections[hubIbcConnIndex], "", "  ")
	fmt.Printf("💈 Hub IBC Connection:\n%s", string(j))

	conn := hubIbcConnections.Connections[hubIbcConnIndex]
	r.SrcConnectionID = conn.ID
	r.SrcClientID = conn.ClientID
	r.DstClientID = conn.Counterparty.ClientID

	return nil
}

func (r *Relayer) HubIbcConnections(hd consts.HubData) (*ConnectionsQueryResult, error) {
	cmd := r.queryConnectionHubCmd(hd)

	hubIbcConnectionsOut, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	var hubIbcConnections ConnectionsQueryResult
	err = json.Unmarshal(hubIbcConnectionsOut.Bytes(), &hubIbcConnections)
	if err != nil {
		return nil, err
	}
	return &hubIbcConnections, nil
}

func (r *Relayer) GetActiveConnectionIDs(
	raData consts.RollappData,
	hd consts.HubData,
) (string, string, error) {
	cmd := r.getQueryRaIbcConnectionsCmd(raData)

	rollappConnectionOutput, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		r.logger.Printf(
			"failed to find connection on the rollapp side for %s: %v",
			r.Rollapp.ID,
			err,
		)
		return "", "", err
	}

	// While there are JSON objects in the stream...
	var rollappIbcConnection ConnectionsQueryResult
	err = json.Unmarshal(rollappConnectionOutput.Bytes(), &rollappIbcConnection)
	if err != nil {
		r.logger.Printf("error while decoding JSON: %v", err)
	}

	if len(rollappIbcConnection.Connections) == 0 {
		r.logger.Printf("no connections found on the rollapp side for %s", r.Rollapp.ID)
		return "", "", nil
	}

	var raActiveConnectionInfo *ConnectionInfo

	j, _ := json.Marshal(rollappIbcConnection.Connections)
	fmt.Printf("\tibc connections:\n%s", string(j))

	for _, conn := range rollappIbcConnection.Connections {
		if conn.State == "STATE_OPEN" {
			raActiveConnectionInfo = &conn
		}
	}
	if raActiveConnectionInfo == nil {
		return "", "", nil
	}
	hubConnectionID := raActiveConnectionInfo.Counterparty.ConnectionID

	// Check if the connection is open on the hub
	hubIbcConnections, err := r.HubIbcConnections(hd)
	if err != nil {
		return "", "", err
	}

	hubConnIndex := slices.IndexFunc(
		hubIbcConnections.Connections, func(conn ConnectionInfo) bool {
			return conn.ID == hubConnectionID
		},
	)

	hubConnection := hubIbcConnections.Connections[hubConnIndex]

	return raActiveConnectionInfo.ID, hubConnection.ID, nil
}

func (r *Relayer) GetActiveConnections(raData consts.RollappData, hd consts.HubData) (
	*ConnectionInfo,
	*ConnectionInfo,
	error,
) {
	rollappConnectionOutput, err := bash.ExecCommandWithStdout(
		r.getQueryRaIbcConnectionsCmd(
			raData,
		),
	)
	if err != nil {
		r.logger.Printf(
			"failed to find connection on the rollapp side for %s: %v",
			r.Rollapp.ID,
			err,
		)
		return nil, nil, err
	}

	// While there are JSON objects in the stream...
	var rollappIbcConnection ConnectionsQueryResult
	err = json.Unmarshal(rollappConnectionOutput.Bytes(), &rollappIbcConnection)
	if err != nil {
		r.logger.Printf("error while decoding JSON: %v", err)
	}

	if len(rollappIbcConnection.Connections) == 0 {
		r.logger.Printf("no connections found on the rollapp side for %s", r.Rollapp.ID)
		return nil, nil, nil
	}

	// TODO: review, why return nil error?
	if rollappIbcConnection.Connections[0].State != "STATE_OPEN" {
		return nil, nil, nil
	}
	hubConnectionID := rollappIbcConnection.Connections[0].Counterparty.ConnectionID

	// Check if the connection is open on the hub
	hubIbcConnections, err := r.HubIbcConnections(hd)
	if err != nil {
		return nil, nil, err
	}

	hubConnIndex := slices.IndexFunc(
		hubIbcConnections.Connections, func(conn ConnectionInfo) bool {
			return conn.ID == hubConnectionID
		},
	)

	hubConnection := hubIbcConnections.Connections[hubConnIndex]

	return &rollappIbcConnection.Connections[0], &hubConnection, nil
}

func (r *Relayer) getQueryRaIbcConnectionsCmd(
	raData consts.RollappData,
) *exec.Cmd {
	args := []string{
		"q",
		"ibc",
		"connection",
		"connections",
		"--node",
		raData.RpcUrl,
		"--chain-id",
		raData.ID,
		"-o", "json",
	}
	cmd := exec.Command(consts.Executables.RollappEVM, args...)
	fmt.Println(cmd.String())

	return cmd
}

// TODO: refactor the limit
func (r *Relayer) queryConnectionHubCmd(hd consts.HubData) *exec.Cmd {
	args := []string{
		"q",
		"ibc",
		"connection",
		"connections",
		"--chain-id",
		hd.ID,
		"--node",
		hd.RpcUrl,
		"-o",
		"json",
		"--limit",
		"100000",
	}

	return exec.Command(consts.Executables.Dymension, args...)
}

func (r *Relayer) queryConnectionsHubCmd() *exec.Cmd {
	args := []string{"q", "connections", r.Hub.ID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}
