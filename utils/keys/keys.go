package keys

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
)

func CreateAddressBinary(
	keyConfig utils.KeyConfig,
	home string,
) (*utils.KeyInfo, error) {
	args := []string{
		"keys", "add", keyConfig.ID, "--keyring-backend", "test",
		"--keyring-dir", filepath.Join(home, keyConfig.Dir),
		"--output", "json",
	}
	createKeyCommand := exec.Command(keyConfig.ChainBinary, args...)
	out, err := bash.ExecCommandWithStdout(createKeyCommand)
	if err != nil {
		return nil, err
	}
	return utils.ParseAddressFromOutput(out)
}
