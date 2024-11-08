package relayer

import (
	"fmt"
	"os"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func VerifyDefaultPath(relayerHome string) (bool, error) {
	pterm.Info.Println("verifying default path")
	cfp := GetConfigFilePath(relayerHome)

	data, err := os.ReadFile(cfp)
	if err != nil {
		pterm.Error.Println("failed to read config file:", err)
		return false, err
	}

	pterm.Info.Println("unmarshalling config file")
	var config map[string]interface{}
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		pterm.Error.Println("failed to unmarshal config file:", err)
		return false, err
	}

	fmt.Println("config", config["paths"])

	pterm.Info.Println("navigating to paths and checking for hub-rollapp")
	// Navigate to paths and check for hub-rollapp
	if paths, ok := config["paths"].(map[interface{}]interface{}); ok {
		if _, exists := paths[consts.DefaultRelayerPath]; exists {
			pterm.Success.Println("hub-rollapp exists in the YAML configuration.")
			return true, nil
		}

		fmt.Println("want:", consts.DefaultRelayerPath)
		fmt.Println("have:", paths[consts.DefaultRelayerPath])

		y, _ := yaml.Marshal(paths)
		fmt.Println("paths", string(y))
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
