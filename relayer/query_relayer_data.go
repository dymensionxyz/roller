package relayer

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

func GetRelayerStatus(config config.RollappConfig) string {
	channels, err := GetChannels(config)
	if err != nil || channels.Src == "" {
		return fmt.Sprintf("Starting...")
	}
	return fmt.Sprintf("Active src, %s <-> %s, dst", channels.Src, channels.Dst)
}

func GetChannels(rollappConfig config.RollappConfig) (ConnectionChannels, error) {
	dstConnectionId, err := GetDstConnectionIDFromYAMLFile(filepath.Join(rollappConfig.Home, consts.ConfigDirName.Relayer,
		"config", "config.yaml"))
	if err != nil {
		return ConnectionChannels{}, err
	}
	if dstConnectionId == "" {
		return ConnectionChannels{}, nil
	}
	connectionChannels, err := GetConnectionChannels(dstConnectionId, rollappConfig)
	if err != nil {
		return ConnectionChannels{}, err
	}
	return connectionChannels, nil
}

// GetDstConnectionIDFromYAMLFile Returns the destination connection ID if it been created already, an empty string otherwise.
func GetDstConnectionIDFromYAMLFile(filename string) (string, error) {

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

type RelayerConfigFile struct {
	Paths map[string]Path `yaml:"paths"`
}

type Path struct {
	Dst Destination `yaml:"dst"`
}

type Destination struct {
	ConnectionID string `yaml:"connection-id"`
}
