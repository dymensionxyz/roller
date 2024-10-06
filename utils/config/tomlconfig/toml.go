package tomlconfig

import (
	"fmt"
	"os"
	"path/filepath"

	naoinatoml "github.com/naoina/toml"
	toml "github.com/pelletier/go-toml"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Write(rlpCfg roller.RollappConfig) error {
	tomlBytes, err := naoinatoml.Marshal(rlpCfg)
	if err != nil {
		return err
	}
	// nolint:gofumpt
	return os.WriteFile(filepath.Join(rlpCfg.Home, consts.RollerConfigFileName), tomlBytes, 0o644)
}

func Load(path string) ([]byte, error) {
	tomlBytes, err := os.ReadFile(path)
	if err != nil {
		return tomlBytes, err
	}

	return tomlBytes, nil
}

func WriteTomlTreeToFile(tomlConfig *toml.Tree, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = file.WriteString(tomlConfig.String())
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func GetKeyFromTomlFile(tmlFilePath, key string) (string, error) {
	tomlTree, err := toml.LoadFile(tmlFilePath)
	if err != nil {
		return "", err
	}
	return tomlTree.Get(key).(string), nil
}

// TODO: improve
func UpdateFieldInToml(tmlFilePath, key string, value any) error {
	tomlCfg, err := toml.LoadFile(tmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", tmlFilePath, err)
	}

	switch v := value.(type) {
	case string, int, int64, float64, bool:
		tomlCfg.Set(key, v)
	default:
		return fmt.Errorf("unsupported type for key %s: %T", key, value)
	}

	return WriteTomlTreeToFile(tomlCfg, tmlFilePath)
}
