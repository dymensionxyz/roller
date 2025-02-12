package eibc

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"
)

func LoadSupportedRollapps(eibcConfigPath string) ([]string, error) {
	data, err := os.ReadFile(eibcConfigPath)
	if err != nil {
		fmt.Println("failed to read: ", err)
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		fmt.Println("failed to unmarshal eibc config file: ", err)
		return nil, err
	}

	if config.Rollapps == nil {
		return []string{}, nil
	}

	keys := make([]string, 0, len(config.Rollapps))
	for k := range config.Rollapps {
		keys = append(keys, k)
	}

	return keys, nil
}

func ReadConfig(eibcConfigPath string) (*Config, error) {
	data, err := os.ReadFile(eibcConfigPath)
	if err != nil {
		pterm.Error.Printf("Error reading file: %v\n", err)
		return nil, err
	}

	// Parse the YAML
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		pterm.Error.Printf("Error reading file: %v\n", err)
		return nil, err
	}

	return &config, nil
}
