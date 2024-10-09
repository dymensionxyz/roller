package upgrades

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/binaries"
)

func NewRollappUpgrade() (*RollappUpgrade, error) {
	ra := RollappUpgrade{
		Name:   "rollapp",
		Binary: consts.Executables.RollappEVM,
	}

	ver, err := ra.Version()
	if err != nil {
		return nil, err
	}
	ra.CurrentVersion = ver

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

	fmt.Println(out.String())

	return out.String(), nil
}
