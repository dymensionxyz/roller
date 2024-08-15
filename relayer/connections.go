package relayer

import (
	"encoding/json"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	roller_utils "github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
)

type ConnectionsQueryResult struct {
	ID           string           `json:"id"`
	ClientID     string           `json:"client_id"`
	Versions     []VersionInfo    `json:"versions"`
	State        string           `json:"state"`
	Counterparty CounterpartyInfo `json:"counterparty"`
	DelayPeriod  string           `json:"delay_period"`
}

type ConnectionQueryResult struct {
	Connection  ConnectionInfo  `json:"connection"`
	Proof       string          `json:"proof"`
	ProofHeight ProofHeightInfo `json:"proof_height"`
}

type ConnectionInfo struct {
	ClientID     string           `json:"client_id"`
	Versions     []VersionInfo    `json:"versions"`
	State        string           `json:"state"`
	Counterparty CounterpartyInfo `json:"counterparty"`
	DelayPeriod  string           `json:"delay_period"`
}

type ProofHeightInfo struct {
	RevisionNumber string `json:"revision_number"`
	RevisionHeight string `json:"revision_height"`
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

func (r *Relayer) GetActiveConnection() (string, error) {
	// try to read connection information from the configuration file
	rlyCfg, err := ReadRlyConfig(r.Home)
	if err != nil {
		return "", err
	}
	connectionIDRollappRaw, err := roller_utils.GetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "dst", "connection-id"},
	)
	if err != nil {
		r.logger.Println("no active rollapp connection id found in the configuration file")
		// return "", err
	}

	connectionIDHubRaw, err := roller_utils.GetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "src", "connection-id"},
	)
	if err != nil {
		r.logger.Println("no active hub connection id found in the configuration file")
		// return "", err
	}

	var connectionIDRollapp string
	if connectionIDRollappRaw != nil {
		//nolint:errcheck
		connectionIDRollapp = connectionIDRollappRaw.(string)
	}

	var connectionIDHub string
	if connectionIDHubRaw != nil {
		//nolint:errcheck
		connectionIDHub = connectionIDHubRaw.(string)
	}

	if connectionIDRollapp == "" || connectionIDHub == "" {
		r.logger.Printf("can't find active connection in the configuration file")
	}
	// END: try to read connection information from the configuration file

	// var hubConnectionInfo ConnectionsQueryResult
	// hubConnectionOutput, err := bash.ExecCommandWithStdout(r.queryConnectionsHubCmd())
	// if err != nil {
	// 	r.logger.Printf("couldn't find any open connections for %s", r.HubID)
	// 	return "", err
	// }

	// fetch connection from the chain
	rollappConnectionOutput, err := bash.ExecCommandWithStdout(
		r.queryConnectionRollappCmd(
			"connection-0",
		),
	)
	if err != nil {
		r.logger.Printf(
			"failed to find connection on the rollapp side for %s: %v",
			r.RollappID,
			err,
		)
		return "", err
	}

	// While there are JSON objects in the stream...
	var outputStruct ConnectionQueryResult
	err = json.Unmarshal(rollappConnectionOutput.Bytes(), &outputStruct)
	if err != nil {
		r.logger.Printf("error while decoding JSON: %v", err)
	}

	// TODO: review, why return nil error?
	if outputStruct.Connection.State != "STATE_OPEN" {
		return "", nil
	}
	hubConnectionID := outputStruct.Connection.Counterparty.ConnectionID

	// Check if the connection is open on the hub
	var res ConnectionQueryResult
	outputHub, err := bash.ExecCommandWithStdout(
		r.queryConnectionHubCmd(hubConnectionID),
	)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(outputHub.Bytes(), &res)
	if err != nil {
		return "", err
	}

	if res.Connection.State != "STATE_OPEN" {
		r.logger.Printf(
			"connection %s is STATE_OPEN on the rollapp, but connection %s is %s on the hub",
			connectionIDRollapp,
			connectionIDHub,
			res.Connection.State,
		)
		return "", nil
	}

	// todo: refactor
	err = roller_utils.SetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "src", "connection-id"},
		hubConnectionID,
	)
	if err != nil {
		return "", err
	}

	err = roller_utils.SetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "dst", "connection-id"},
		"connection-0",
	)
	if err != nil {
		return "", err
	}

	err = roller_utils.SetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "src", "client-id"},
		outputStruct.Connection.Counterparty.ClientID,
	)
	if err != nil {
		return "", err
	}

	err = roller_utils.SetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "src", "client-id"},
		"07-tendermint-0",
	)
	if err != nil {
		return "", err
	}

	return "connection-0", nil
}

func (r *Relayer) queryConnectionRollappCmd(connectionID string) *exec.Cmd {
	args := []string{"q", "connection", r.RollappID, connectionID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) queryConnectionHubCmd(connectionID string) *exec.Cmd {
	args := []string{"q", "connection", r.HubID, connectionID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) queryConnectionsHubCmd() *exec.Cmd {
	args := []string{"q", "connections", r.HubID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}
