package start

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type RelayerConfigFile struct {
	Paths map[string]Path `yaml:"paths"`
}

type Path struct {
	Dst Destination `yaml:"dst"`
}

type Destination struct {
	ConnectionID string `yaml:"connection-id"`
}

func GetDstConnectionIDFromYAMLFile(filename string) (string, error) {
	/**
	Returns the destination connection ID if it been created already, an empty string otherwise.
	*/
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return "", err
	}
	var config RelayerConfigFile
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return "", err
	}
	for _, path := range config.Paths {
		return path.Dst.ConnectionID, nil
	}
	return "", fmt.Errorf("No paths found in YAML data")
}
