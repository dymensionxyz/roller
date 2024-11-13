package eibc

import (
	"os/exec"
	"strconv"

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
		strconv.Itoa(consts.DefaultTxFee),
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
	eibcHome, metadata, policyDefinitionFilePath string,
	hd consts.HubData,
) *exec.Cmd {
	cmd := exec.Command(
		consts.Executables.Dymension,
		"tx",
		"group",
		"create-group-policy",
		consts.KeysIds.Eibc,
		"1",
		metadata,
		policyDefinitionFilePath,
		"--fees",
		strconv.Itoa(consts.DefaultTxFee),
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
