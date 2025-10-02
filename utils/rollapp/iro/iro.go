package iro

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
)

func IsTokenGraduates(raID string, hd consts.HubData) bool {
	cmd := exec.Command(consts.Executables.Dymension, "q", "iro", "plan-by-rollapp", raID,
		"--output",
		"json",
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
	)

	fmt.Println(cmd.String())

	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return false
	}

	fmt.Println(out.String())

	var resp PlanResponse
	if err := json.Unmarshal(out.Bytes(), &resp); err != nil {
		return false
	}

	j, _ := json.Marshal(resp)
	fmt.Println(string(j))

	isGraduated, err := strconv.Atoi(resp.GraduatedPoolID)
	if err != nil {
		return false
	}
	if isGraduated > 0 {
		return true
	}

	return false
}

type PlanResponse struct {
	plan Plan `json:"plan"`
}

type Plan struct {
	ID              string `json:"id"`
	RollappID       string `json:"rollapp_id"`
	GraduatedPoolID string `json:"graduated_pool_id"`
}
