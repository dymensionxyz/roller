package utils

import (
	"github.com/pelletier/go-toml"
	"io/ioutil"
	"path/filepath"
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

type InitConfig struct {
	Home          string
	RollappID     string
	RollappBinary string
	Denom         string
	Decimals      uint64
	HubData       HubData
}

const RollerConfigFileName = "config.toml"

type HubData = struct {
	API_URL string
	ID      string
	RPC_URL string
}
