package eibc

import (
	"encoding/json"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func GetGrantsByGrantee(policyAddr string, hd consts.HubData) (*GrantsByGranteeResponse, error) {
	cmd := GetGrantsByGranteeCmd(policyAddr, hd)

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, err
	}

	var grantsByGrantee GrantsByGranteeResponse
	err = json.Unmarshal(out.Bytes(), &grantsByGrantee)
	if err != nil {
		return nil, err
	}

	return &grantsByGrantee, err
}

func GetGrantsByGranteeCmd(policyAddr string, hd consts.HubData) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"q",
		"authz",
		"grants-by-grantee",
		policyAddr,
		"-o",
		"json",
		"--node", hd.RpcUrl,
	)

	return cmd
}

type GrantsByGranteeResponse struct {
	Grants []Grant `json:"grants"`
}

type Grant struct {
	Granter       string        `json:"granter"`
	Grantee       string        `json:"grantee"`
	Authorization Authorization `json:"authorization"`
}

type Authorization struct {
	Type  string    `json:"type"`
	Value AuthValue `json:"value"`
}

type AuthValue struct {
	Rollapps []Rollapp `json:"rollapps"`
}

type Rollapp struct {
	RollappID           string   `json:"rollapp_id"`
	Denoms              []string `json:"denoms"`
	MaxPrice            []Amount `json:"max_price"`
	MinFeePercentage    string   `json:"min_fee_percentage"`
	OperatorFeeShare    string   `json:"operator_fee_share"`
	SettlementValidated bool     `json:"settlement_validated"`
	SpendLimit          []Amount `json:"spend_limit"`
}

type Amount struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}
