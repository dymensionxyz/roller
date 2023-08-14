package utils

import (
	"fmt"
	"github.com/pelletier/go-toml"
	"os"
)

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

func UpdateFieldInToml(tmlFilePath, key, value string) error {
	var tomlCfg, err = toml.LoadFile(tmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", tmlFilePath, err)
	}
	tomlCfg.Set(key, value)
	return WriteTomlTreeToFile(tomlCfg, tmlFilePath)
}
