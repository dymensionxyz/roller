package solana

import (
	"os"
	"path/filepath"

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

func loadConfigFromTOML(filePath string) Solana {
	var config Solana
	data, err := os.ReadFile(filePath)
	if err != nil {
		return Solana{}
	}

	err = toml.Unmarshal(data, &config)
	if err != nil {
		return Solana{}
	}

	return config
}

func GetCfgFilePath(rollerHome string) string {
	return filepath.Join(
		rollerHome,
		"da-light-client",
		"config.toml",
	)
}
