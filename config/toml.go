package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v3"
)

func WriteConfigToTOML(rlpCfg RollappConfig) error {
	tomlBytes, err := toml.Marshal(rlpCfg)
	if err != nil {
		return err
	}
	// nolint:gofumpt
	return os.WriteFile(filepath.Join(rlpCfg.Home, RollerConfigFileName), tomlBytes, 0644)
}

// TODO: should be called from root command
func LoadConfigFromTOML(root string) (RollappConfig, error) {
	var config RollappConfig
	tomlBytes, err := os.ReadFile(filepath.Join(root, RollerConfigFileName))
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

// TODO: related to:
// https://github.com/dymensionxyz/eibc-client/issues/22
// eibc config should be moved to toml format so
// all config files use the same file extension
func LoadConfigFromYAML(root string) (RollappConfig, error) {
	var config RollappConfig
	yamlBytes, err := os.ReadFile(filepath.Join(root, RollerConfigFileName))
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
