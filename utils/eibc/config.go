package eibc

import (
	"fmt"
	"os"

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
