package initconfig

import (
	"io/ioutil"
	"path/filepath"

	"github.com/pelletier/go-toml"
)

func WriteConfigToTOML(InitConfig InitConfig) error {
	tomlBytes, err := toml.Marshal(InitConfig)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filepath.Join(InitConfig.Home, RollerConfigFileName), tomlBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfigFromTOML(root string) (InitConfig, error) {
	var config InitConfig
	tomlBytes, err := ioutil.ReadFile(filepath.Join(root, RollerConfigFileName))
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
