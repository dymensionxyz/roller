package relayer

import (
	"encoding/json"
	"fmt"

	"github.com/dymensionxyz/roller/cmd/utils"
)

type RollappConnectionQueryResult struct {
	ID           string           `json:"id"`
	ClientID     string           `json:"client_id"`
	Versions     []VersionInfo    `json:"versions"`
	State        string           `json:"state"`
	Counterparty CounterpartyInfo `json:"counterparty"`
	DelayPeriod  string           `json:"delay_period"`
}

type HubConnectionQueryResult struct {
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

func (r *Relayer) IsLatestConnectionOpen() (bool, string, error) {
	output, err := utils.ExecBashCommandWithStdout(r.queryConnectionsRollappCmd())
	if err != nil {
		return false, "", err
	}

	if output.Len() == 0 {
		return false, "", nil
	}

	// While there are JSON objects in the stream...
	var outputStruct RollappConnectionQueryResult

	dec := json.NewDecoder(&output)
	for dec.More() {
		err = dec.Decode(&outputStruct)
		if err != nil {
			return false, "", fmt.Errorf("error while decoding JSON: %v", err)
		}
	}

	if outputStruct.State != "STATE_OPEN" {
		return false, "", nil
	}

	// Check if the connection is open on the hub
	var res HubConnectionQueryResult
	outputHub, err := utils.ExecBashCommandWithStdout(r.queryConnectionsHubCmd(outputStruct.Counterparty.ConnectionID))
	if err != nil {
		return false, "", err
	}
	err = json.Unmarshal(outputHub.Bytes(), &res)
	if err != nil {
		return false, "", err
	}

	if res.Connection.State != "STATE_OPEN" {
		r.logger.Printf("connection %s is STATE_OPEN on the rollapp, but connection %s is %s on the hub",
			outputStruct.ID, outputStruct.Counterparty.ConnectionID, res.Connection.State,
		)
		return false, "", nil
	}

	return true, outputStruct.ID, nil
}
