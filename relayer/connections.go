package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	roller_utils "github.com/dymensionxyz/roller/utils"
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
	rlyCfg, err := ReadRlyConfig(r.Home)
	if err != nil {
		return "", err
	}
	connectionIDRollapp_raw, err := roller_utils.GetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "dst", "connection-id"},
	)
	if err != nil {
		return "", err
	}

	connectionIDHub_raw, err := roller_utils.GetNestedValue(
		rlyCfg,
		[]string{"paths", consts.DefaultRelayerPath, "src", "connection-id"},
	)
	if err != nil {
		return "", err
	}

	//nolint:errcheck
	connectionIDRollapp := connectionIDRollapp_raw.(string)
	//nolint:errcheck
	connectionIDHub := connectionIDHub_raw.(string)

	if connectionIDRollapp == "" || connectionIDHub == "" {
		r.logger.Printf("can't find active connection in the config")
		return "", nil
	}

	output, err := utils.ExecBashCommandWithStdout(r.queryConnectionRollappCmd(connectionIDRollapp))
	if err != nil {
		return "", err
	}

	// While there are JSON objects in the stream...
	var outputStruct ConnectionQueryResult

	dec := json.NewDecoder(&output)
	for dec.More() {
		err = dec.Decode(&outputStruct)
		if err != nil {
			return "", fmt.Errorf("error while decoding JSON: %v", err)
		}
	}

	if outputStruct.Connection.State != "STATE_OPEN" {
		return "", nil
	}

	// Check if the connection is open on the hub
	var res ConnectionQueryResult
	outputHub, err := utils.ExecBashCommandWithStdout(r.queryConnectionsHubCmd(connectionIDHub))
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

	return connectionIDRollapp, nil
}

func (r *Relayer) queryConnectionRollappCmd(connectionID string) *exec.Cmd {
	args := []string{"q", "connection", r.RollappID, connectionID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}

func (r *Relayer) queryConnectionsHubCmd(connectionID string) *exec.Cmd {
	args := []string{"q", "connection", r.HubID, connectionID}
	args = append(args, "--home", filepath.Join(r.Home, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}
