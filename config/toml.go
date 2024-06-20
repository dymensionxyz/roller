package config

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"
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
