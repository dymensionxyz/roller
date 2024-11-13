package eibc

import (
	"encoding/json"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func GetPolicies(home string, hd consts.HubData) (*GetGroupPoliciesResponse, error) {
	cmd := GetPoliciesCmd(home, hd)
	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	// nolint:errcheck
	err = cmd.Wait()
	if err != nil {
		return nil, err
	}

	var outResp GetGroupPoliciesResponse
	err = json.Unmarshal(out.Bytes(), &outResp)
	if err != nil {
		return nil, err
	}

	return &outResp, nil
}

func GetPoliciesCmd(home string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"group",
		"group-policies-by-group",
		"1",
		"-o",
		"json",
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
		"--home",
		home,
	)

	return cmd
}

type GetGroupPoliciesResponse struct {
	GroupPolicies []GroupPolicy `json:"group_policies,omitempty"`
	Pagination    struct {
		NextKey interface{} `json:"next_key,omitempty"`
		Total   string      `json:"total,omitempty"`
	} `json:"pagination,omitempty"`
}

type GroupPolicy struct {
	Address        string         `json:"address,omitempty"`
	Admin          string         `json:"admin,omitempty"`
	CreatedAt      string         `json:"created_at,omitempty"`
	DecisionPolicy DecisionPolicy `json:"decision_policy,omitempty"`
	GroupID        string         `json:"group_id,omitempty"`
	Metadata       string         `json:"metadata,omitempty"`
	Version        string         `json:"version,omitempty"`
}

type DecisionPolicy struct {
	Type       string `json:"@type,omitempty"`
	Percentage string `json:"percentage,omitempty"`
	Windows    Window `json:"windows,omitempty"`
}

type Window struct {
	MinExecutionPeriod string `json:"min_execution_period,omitempty"`
	VotingPeriod       string `json:"voting_period,omitempty"`
}
