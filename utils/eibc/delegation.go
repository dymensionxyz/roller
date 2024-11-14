package eibc

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func GetCreateGroupDelegationCmd(
	eibcHome, metadata, membersDefinitionFilePath string,
	hd consts.HubData,
) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx",
		"group",
		"create-group",
		consts.KeysIds.Eibc,
		"--keyring-backend",
		"test",
		metadata,
		membersDefinitionFilePath,
		"--fees",
		fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
		"-y",
		"--home",
		eibcHome,
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
	)
	return cmd
}

func GetCreateGroupPolicyCmd(
	eibcHome, metadata, policyDefinitionFilePath, groupID string,
	hd consts.HubData,
) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx",
		"group",
		"create-group-policy",
		consts.KeysIds.Eibc,
		groupID,
		metadata,
		policyDefinitionFilePath,
		"--fees",
		fmt.Sprintf("%d%s", consts.DefaultTxFee, consts.Denoms.Hub),
		"-y",
		"--home",
		eibcHome,
		"--keyring-backend",
		"test",
		"--node",
		hd.RpcUrl,
		"--chain-id",
		hd.ID,
	)

	return cmd
}
