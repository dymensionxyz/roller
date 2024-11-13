package eibc

import (
	"encoding/json"
	"os/exec"
	"time"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func GetGroups(admin string, hd consts.HubData) (*GroupsResponse, error) {
	cmd := GetGroupsCmd(admin, hd)
	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	var groups GroupsResponse
	err = json.Unmarshal(out.Bytes(), &groups)
	if err != nil {
		return nil, err
	}

	return &groups, err
}

// dymd q group groups-by-admin dym1tqgc0zdr85t2l3a4lgppcduk7rsnfdtrvafct0 -o json
func GetGroupsCmd(addr string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"group",
		"groups-by-admin",
		addr,
		"-o",
		"json",
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
	)

	return cmd
}

// Group represents a single group in the "groups" array
type Group struct {
	ID          string    `json:"id"`
	Admin       string    `json:"admin"`
	Metadata    string    `json:"metadata"`
	Version     string    `json:"version"`
	TotalWeight string    `json:"total_weight"`
	CreatedAt   time.Time `json:"created_at"`
}

// GroupsResponse represents the auth groups returned by `dymd q group groups`
type GroupsResponse struct {
	Groups []Group `json:"groups"`
}
