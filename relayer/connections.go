package relayer

import (
	"encoding/json"
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

func (r *Relayer) GetActiveConnectionIDs(
	raData consts.RollappData,
	hd consts.HubData,
) (string, string, error) {
	rollappConnectionOutput, err := bash.ExecCommandWithStdout(
		r.queryConnectionRollappCmd(
			raData,
		),
	)
	if err != nil {
		r.logger.Printf(
			"failed to find connection on the rollapp side for %s: %v",
			r.RollappID,
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
		r.logger.Printf("no connections found on the rollapp side for %s", r.RollappID)
		return "", "", nil
	}

	// TODO: review, why return nil error?
	if rollappIbcConnection.Connections[0].State != "STATE_OPEN" {
		return "", "", nil
	}
	hubConnectionID := rollappIbcConnection.Connections[0].Counterparty.ConnectionID

	// Check if the connection is open on the hub
	var hubIbcConnection ConnectionsQueryResult
	outputHub, err := bash.ExecCommandWithStdout(
		r.queryConnectionHubCmd(hd),
	)
	if err != nil {
		return "", "", err
	}

	err = json.Unmarshal(outputHub.Bytes(), &hubIbcConnection)
	if err != nil {
		return "", "", err
	}

	hubConnIndex := slices.IndexFunc(
		hubIbcConnection.Connections, func(conn ConnectionInfo) bool {
			return conn.ID == hubConnectionID
		},
	)

	hubConnection := hubIbcConnection.Connections[hubConnIndex]

	return rollappIbcConnection.Connections[0].ID, hubConnection.ID, nil
}

func (r *Relayer) GetActiveConnections(raData consts.RollappData, hd consts.HubData) (
	*ConnectionInfo,
	*ConnectionInfo,
	error,
) {
	rollappConnectionOutput, err := bash.ExecCommandWithStdout(
		r.queryConnectionRollappCmd(
			raData,
		),
	)
	if err != nil {
		r.logger.Printf(
			"failed to find connection on the rollapp side for %s: %v",
			r.RollappID,
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
		r.logger.Printf("no connections found on the rollapp side for %s", r.RollappID)
		return nil, nil, nil
	}

	// TODO: review, why return nil error?
	if rollappIbcConnection.Connections[0].State != "STATE_OPEN" {
		return nil, nil, nil
	}
	hubConnectionID := rollappIbcConnection.Connections[0].Counterparty.ConnectionID

	// Check if the connection is open on the hub
	var hubIbcConnection ConnectionsQueryResult
	outputHub, err := bash.ExecCommandWithStdout(
		r.queryConnectionHubCmd(hd),
	)
	if err != nil {
		return nil, nil, err
	}

	err = json.Unmarshal(outputHub.Bytes(), &hubIbcConnection)
	if err != nil {
		return nil, nil, err
	}

	hubConnIndex := slices.IndexFunc(
		hubIbcConnection.Connections, func(conn ConnectionInfo) bool {
			return conn.ID == hubConnectionID
		},
	)

	hubConnection := hubIbcConnection.Connections[hubConnIndex]

	return &rollappIbcConnection.Connections[0], &hubConnection, nil
}

func (r *Relayer) queryConnectionRollappCmd(
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

	return cmd
}

func (r *Relayer) queryConnectionHubCmd(hd consts.HubData) *exec.Cmd {
	args := []string{
		"q",
		"ibc",
		"connection",
		"connections",
		"--chain-id",
		hd.ID,
		"--node",
		hd.RPC_URL,
		"-o",
		"json",
	}

	return exec.Command(consts.Executables.Dymension, args...)
}

func (r *Relayer) queryConnectionsHubCmd() *exec.Cmd {
	args := []string{"q", "connections", r.HubID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}
