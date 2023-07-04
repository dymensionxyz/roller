package config

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

func WriteConfigToTOML(InitConfig RollappConfig) error {
	tomlBytes, err := toml.Marshal(InitConfig)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(InitConfig.Home, RollerConfigFileName), tomlBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

// TODO: should be called from root command
func LoadConfigFromTOML(root string) (RollappConfig, error) {
	var config RollappConfig
	tomlBytes, err := ioutil.ReadFile(filepath.Join(root, RollerConfigFileName))
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
