package eibc

import (
	"encoding/json"
	"os"
)

func LoadSupportedRollapps(eibcConfigPath string) ([]string, error) {
	data, err := os.ReadFile(eibcConfigPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
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
