package relayer

import (
	"os"

	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func VerifyDefaultPath(relayerHome string) (bool, error) {
	cfp := GetConfigFilePath(relayerHome)

	data, err := os.ReadFile(cfp)
	if err != nil {
		pterm.Error.Println("failed to read config file:", err)
		return false, err
	}

	var config Config
	err = yaml.Unmarshal(data, &config)
	if err != nil {
		pterm.Error.Println("failed to unmarshal config file:", err)
		return false, err
	}

	if config.Paths == nil || config.Paths.HubRollapp == nil {
		pterm.Error.Println("hub-rollapp not found in the YAML configuration.")
		return false, nil
	}

	return true, nil
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
