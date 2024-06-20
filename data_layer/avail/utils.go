package avail

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func writeConfigToTOML(path string, c Avail) error {
	tomlBytes, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	// nolint:gofumpt
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	// nolint:gofumpt
	err = os.WriteFile(path, tomlBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func loadConfigFromTOML(path string) (Avail, error) {
	var config Avail
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
