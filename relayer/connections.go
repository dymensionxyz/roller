package relayer

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
)

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

func (r *Relayer) RaIbcConnections(
	raData consts.RollappData,
) (*ConnectionsQueryResult, error) {
	cmd := r.getQueryRaIbcConnectionsCmd(raData)

	raConnectionsOut, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		r.logger.Printf(
			"failed to find connection on the rollapp side for %s: %v",
			r.Rollapp.ID,
			err,
		)
		return nil, err
	}

	var raIbcConnections ConnectionsQueryResult
	err = json.Unmarshal(raConnectionsOut.Bytes(), &raIbcConnections)
	if err != nil {
		return nil, err
	}
	return &raIbcConnections, nil
}

func (r *Relayer) UpdateDefaultPath() error {
	updates := map[string]interface{}{
		// hub
		fmt.Sprintf("paths.%s.src.client-id", consts.DefaultRelayerPath):     r.SrcClientID,
		fmt.Sprintf("paths.%s.src.connection-id", consts.DefaultRelayerPath): r.SrcConnectionID,

		// ra
		fmt.Sprintf("paths.%s.dst.client-id", consts.DefaultRelayerPath):     r.DstClientID,
		fmt.Sprintf("paths.%s.dst.connection-id", consts.DefaultRelayerPath): r.DstConnectionID,
	}
	err := yamlconfig.UpdateNestedYAML(r.ConfigFilePath, updates)
	if err != nil {
		return err
	}

	return nil
}

func (r *Relayer) ConnectionInfoFromRaConnID(
	raData consts.RollappData,
	raIbcConnectionID string,
) error {
	raIbcConnections, err := r.RaIbcConnections(raData)
	if err != nil {
		return err
	}

	if len(raIbcConnections.Connections) == 0 {
		return fmt.Errorf("no connections found on the rollapp side for %s", r.Rollapp.ID)
	}

	raIbcConnIndex := slices.IndexFunc(
		raIbcConnections.Connections, func(ibcConn ConnectionInfo) bool {
			return ibcConn.ID == raIbcConnectionID && ibcConn.State == "STATE_OPEN"
		},
	)

	if raIbcConnIndex == -1 {
		return fmt.Errorf("no open channel found for %s", r.Rollapp.ID)
	}

	j, _ := json.MarshalIndent(raIbcConnections.Connections[raIbcConnIndex], "", "  ")
	fmt.Printf("💈 Hub IBC Connection:\n%s", string(j))

	conn := raIbcConnections.Connections[raIbcConnIndex]
	r.SrcConnectionID = conn.Counterparty.ConnectionID
	r.SrcClientID = conn.Counterparty.ClientID
	r.DstClientID = conn.ClientID

	return nil
}

func (r *Relayer) GetActiveConnectionIDs(
	raData consts.RollappData,
	hd consts.HubData,
) (string, string, error) {
	// While there are JSON objects in the stream...
	raIbcConnections, err := r.RaIbcConnections(
		raData,
	)
	if err != nil {
		return "", "", err
	}

	if len(raIbcConnections.Connections) == 0 {
		r.logger.Printf("no connections found on the rollapp side for %s", r.Rollapp.ID)
		return "", "", nil
	}

	var raActiveConnectionInfo *ConnectionInfo

	for _, conn := range raIbcConnections.Connections {
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
	args = append(args, "--home", filepath.Join(r.RollerHome, consts.ConfigDirName.Relayer))
	return exec.Command(consts.Executables.Relayer, args...)
}
