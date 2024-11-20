package upgrades

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/binaries"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/migrations"
)

func NewRollappUpgrade(vmType string) (*RollappUpgrade, error) {
	ra := RollappUpgrade{
		RollappType: vmType,
		Software: Software{
			Name:   "rollapp",
			Binary: consts.Executables.RollappEVM,
		},
	}

	ver, err := ra.Version()
	if err != nil {
		return nil, err
	}
	ra.CurrentVersion = ver

	verCom, err := ra.VersionCommit()
	if err != nil {
		return nil, err
	}
	ra.CurrentVersionCommit = verCom

	return &ra, nil
}

func (ra RollappUpgrade) Version() (string, error) {
	isAvailable := binaries.IsAvailable(consts.Executables.RollappEVM)
	if !isAvailable {
		return "", exec.ErrNotFound
	}

	cmd := exec.Command(ra.Binary, "version")
	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", err
	}

	ver := out.String()
	ver = strings.TrimSpace(ver)

	return ver, nil
}

func (ra RollappUpgrade) VersionCommit() (string, error) {
	isAvailable := binaries.IsAvailable(consts.Executables.RollappEVM)
	if !isAvailable {
		return "", exec.ErrNotFound
	}

	var commit string

	commit, _ = dependencies.ExtractCommitFromBinaryVersion(consts.Executables.RollappEVM)
	if commit == "" {
		version, err := ra.Version()
		if err != nil {
			return "", err
		}

		commit, err = migrations.GetCommitFromTag(
			"dymensionxyz",
			fmt.Sprintf("rollapp-%s", ra.RollappType),
			version,
		)
		if err != nil {
			return "", err
		}

	}

	return commit, nil
}
