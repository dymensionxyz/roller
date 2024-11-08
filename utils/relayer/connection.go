package relayer

import (
	"encoding/json"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func VerifyDefaultPath(relayerHome string) (bool, error) {
	cfp := GetConfigFilePath(relayerHome)

	data, err := os.ReadFile(cfp)
	if err != nil {
		return false, err
	}

	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return false, err
	}

	j, _ := json.MarshalIndent(config, "", "  ")
	fmt.Println(string(j))

	// Navigate to paths and check for hub-rollapp
	if paths, ok := config["paths"].(map[interface{}]interface{}); ok {
		if _, exists := paths[consts.DefaultRelayerPath]; exists {
			fmt.Println("hub-rollapp exists in the YAML configuration.")
			return true, nil
		}
	}

	return false, nil
}

func VerifyPathSrcChain(relayerHome string, hd consts.HubData) (bool, error) {
	cfp := GetConfigFilePath(relayerHome)

	data, err := os.ReadFile(cfp)
	if err != nil {
		return false, err
	}

	// Unmarshal the YAML into the Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return false, err
	}

	// Check if src.chain-id has a specific value
	if config.Paths.HubRollapp.Src.ChainID == hd.ID {
		return true, nil
	}

	return false, nil
}

func VerifyPathDstChain(relayerHome, raID string) (bool, error) {
	cfp := GetConfigFilePath(relayerHome)

	data, err := os.ReadFile(cfp)
	if err != nil {
		return false, err
	}

	// Unmarshal the YAML into the Config struct
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		return false, err
	}

	// Check if src.chain-id has a specific value
	if config.Paths.HubRollapp.Dst.ChainID == raID {
		return true, nil
	}

	return false, nil
}
