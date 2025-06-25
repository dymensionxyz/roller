package ethereum

import (
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func writeConfigToTOML(path string, e Ethereum) error {
	tomlBytes, err := toml.Marshal(e)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	err = os.WriteFile(path, tomlBytes, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func loadConfigFromTOML(path string) (Ethereum, error) {
	var config Ethereum
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
