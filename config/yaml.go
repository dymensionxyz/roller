package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

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
