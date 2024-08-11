package tomlconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	dymensionratypes "github.com/dymensionxyz/dymension/v3/x/rollapp/types"
	naoinatoml "github.com/naoina/toml"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
)

func Write(rlpCfg config.RollappConfig) error {
	tomlBytes, err := naoinatoml.Marshal(rlpCfg)
	if err != nil {
		return err
	}
	// nolint:gofumpt
	return os.WriteFile(filepath.Join(rlpCfg.Home, consts.RollerConfigFileName), tomlBytes, 0o644)
}

// TODO: should be called from root command
func LoadRollerConfig(root string) (config.RollappConfig, error) {
	var config config.RollappConfig
	tomlBytes, err := os.ReadFile(filepath.Join(root, consts.RollerConfigFileName))
	if err != nil {
		return config, err
	}
	err = naoinatoml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func Load(path string) ([]byte, error) {
	tomlBytes, err := os.ReadFile(path)
	if err != nil {
		return tomlBytes, err
	}

	return tomlBytes, nil
}

func LoadRollappMetadataFromChain(home, raID string) (*config.RollappConfig, error) {
	var config config.RollappConfig
	var ra dymensionratypes.QueryGetRollappResponse

	tomlBytes, err := os.ReadFile(filepath.Join(home, consts.RollerConfigFileName))
	if err != nil {
		return &config, err
	}
	err = naoinatoml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return &config, err
	}

	getRollappCmd := exec.Command(
		consts.Executables.Dymension,
		"q", "rollapp", "show",
		raID,
	)

	out, err := bash.ExecCommandWithStdout(getRollappCmd)
	if err != nil {
		return &config, err
	}

	err = json.Unmarshal(out.Bytes(), &ra)
	if err != nil {
		return &config, err
	}

	j, _ := json.MarshalIndent(ra, "", "  ")
	fmt.Println(string(j))

	return &config, nil
}
