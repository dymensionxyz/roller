package solana

import (
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/pelletier/go-toml/v2"
)

func writeConfigToTOML(filePath string, config Solana) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	// Marshal config to TOML
	data, err := toml.Marshal(config)
	if err != nil {
		return err
	}

	// Write to file
	return os.WriteFile(filePath, data, 0o644)
}

func loadConfigFromTOML(path string) (Solana, error) {
	var config Solana
	tomlBytes, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}

func GetCfgFilePath(rollerHome string) string {
	return filepath.Join(rollerHome, consts.ConfigDirName.DALightNode, ConfigFileName)
}
