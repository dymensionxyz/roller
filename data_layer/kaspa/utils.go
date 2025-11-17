package kaspa

import (
	"errors"
	"os"
	"path/filepath"

	"github.com/pelletier/go-toml"

	"github.com/dymensionxyz/roller/cmd/consts"
)

const mnemonicEnvFileName = "kaspa_mnemonic.env"

func writeConfigToTOML(path string, c Kaspa) error {
	tomlBytes, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	dir := filepath.Dir(path)
	// nolint:gofumpt
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	// nolint:gofumpt
	err = os.WriteFile(path, tomlBytes, 0o644)
	if err != nil {
		return err
	}

	return nil
}

func loadConfigFromTOML(path string) (Kaspa, error) {
	var config Kaspa
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

func GetMnemonicEnvFilePath(rollerHome string) string {
	return filepath.Join(rollerHome, consts.ConfigDirName.DALightNode, mnemonicEnvFileName)
}

func EnsureMnemonicEnvFile(rollerHome string) (string, error) {
	envPath := GetMnemonicEnvFilePath(rollerHome)
	if err := os.MkdirAll(filepath.Dir(envPath), 0o755); err != nil {
		return "", err
	}
	_, err := os.Stat(envPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			if err := os.WriteFile(envPath, []byte{}, 0o600); err != nil {
				return "", err
			}
		} else {
			return "", err
		}
	}
	return envPath, nil
}
