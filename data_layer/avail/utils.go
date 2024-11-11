package avail

import (
	"fmt"
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

func loadConfigFromTOML(path string) (Avail, error) {
	var config Avail
	tomlBytes, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	fmt.Println("unmarshelling error here.......", err)
	if err != nil {
		return config, err
	}

	config.Mnemonic = "bottom drive obey lake curtain smoke basket hold race lonely fit walk//Alice" // TODO : fix me

	fmt.Println("config path and config.......", path, config)

	return config, nil
}

func GetCfgFilePath(rollerHome string) string {
	return filepath.Join(rollerHome, consts.ConfigDirName.DALightNode, ConfigFileName)
}
