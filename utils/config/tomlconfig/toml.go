package tomlconfig

import (
	"fmt"
	"os"

	"github.com/pelletier/go-toml"
)

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
	// nolint: errcheck
	defer file.Close()

	_, err = file.WriteString(tomlConfig.String())
	if err != nil {
		return err
	}
	return nil
}

func GetKeyFromFile(tmlFilePath, key string) (string, error) {
	tomlTree, err := toml.LoadFile(tmlFilePath)
	if err != nil {
		return "", err
	}

	return tomlTree.Get(key).(string), nil
}

// TODO: improve
func UpdateFieldInFile(tmlFilePath, key string, value any) error {
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

func RemoveFieldFromFile(tmlFilePath, keyPath string) error {
	tomlCfg, err := toml.LoadFile(tmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", tmlFilePath, err)
	}

	if !tomlCfg.Has(keyPath) {
		return fmt.Errorf("key %s does not exist", keyPath)
	}

	err = tomlCfg.Delete(keyPath)
	if err != nil {
		return err
	}

	return WriteTomlTreeToFile(tomlCfg, tmlFilePath)
}

func ReplaceFieldInFile(tmlFilePath, oldPath, newPath string, value any) error {
	tomlCfg, err := toml.LoadFile(tmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", tmlFilePath, err)
	}

	if !tomlCfg.Has(oldPath) {
		return fmt.Errorf("old key %s does not exist", oldPath)
	}

	var writeableValue any
	if value == nil {
		writeableValue = tomlCfg.Get(oldPath)
	} else {
		writeableValue = value
	}

	err = tomlCfg.Delete(oldPath)
	if err != nil {
		return err
	}

	switch v := writeableValue.(type) {
	case string, int, int64, float64, bool:
		tomlCfg.Set(newPath, v)
	default:
		return fmt.Errorf("unsupported type for new key %s: %T", newPath, value)
	}

	return WriteTomlTreeToFile(tomlCfg, tmlFilePath)
}
