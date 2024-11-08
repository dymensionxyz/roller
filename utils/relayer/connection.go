package relayer

import (
	"fmt"
	"os"
	"reflect"

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
	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		pterm.Error.Println("failed to unmarshal config file:", err)
		return false, err
	}

	pterm.Info.Println("navigating to paths and checking for hub-rollapp")
	// Navigate to paths and check for hub-rollapp

	if !reflect.DeepEqual(config, Config{}) {
		fmt.Println("want:", consts.DefaultRelayerPath)
		fmt.Println("have:", config.Paths.HubRollapp)

		y, _ := yaml.Marshal(config.Paths)
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
