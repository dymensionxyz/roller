package keys

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"

	"github.com/dymensionxyz/roller/utils/bash"
)

func ParseAddressFromOutput(output bytes.Buffer) (*KeyInfo, error) {
	key := &KeyInfo{}
	err := json.Unmarshal(output.Bytes(), key)
	if err != nil {
		return nil, err
	}
	return key, nil
}

func CreateAddressBinary(
	keyConfig KeyConfig,
	home string,
) (*KeyInfo, error) {
	args := []string{
		"keys", "add", keyConfig.ID, "--keyring-backend", "test",
		"--keyring-dir", filepath.Join(home, keyConfig.Dir),
		"--output", "json",
	}

	if keyConfig.ShouldRecover {
		args = append(args, "--recover")
	}
	createKeyCommand := exec.Command(keyConfig.ChainBinary, args...)
	out, err := bash.ExecCommandWithStdout(createKeyCommand)
	if err != nil {
		return nil, err
	}
	return ParseAddressFromOutput(out)
}

func IsAddressWithNameInKeyring(
	info KeyConfig,
	home string,
) (bool, error) {
	keyringDir := filepath.Join(home, info.Dir)

	cmd := exec.Command(
		info.ChainBinary,
		"keys", "list", "--output", "json",
		"--keyring-backend", "test", "--keyring-dir", keyringDir,
	)

	var ki []KeyInfo
	out, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return false, err
	}

	fmt.Println(out.String())

	err = json.Unmarshal(out.Bytes(), &ki)
	if err != nil {
		return false, err
	}

	if len(ki) == 0 {
		return false, nil
	}

	return slices.ContainsFunc(
		ki, func(i KeyInfo) bool {
			return strings.EqualFold(i.Name, info.ID)
		},
	), nil
}
