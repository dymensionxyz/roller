package relayer

import (
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
)

func UpdateConfigWithDefaultValues(relayerHome string, rollerData roller.RollappConfig) error {
	pterm.Info.Println("updating application relayer config")
	relayerConfigPath := filepath.Join(relayerHome, "config", "config.yaml")
	updates := map[string]interface{}{
		fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.HubData.ID): 1.5,
		fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.RollappID):  1.3,
		fmt.Sprintf("chains.%s.value.is-dym-hub", rollerData.HubData.ID):     true,
		fmt.Sprintf(
			"chains.%s.value.http-addr",
			rollerData.HubData.ID,
		): rollerData.HubData.API_URL,
		fmt.Sprintf("chains.%s.value.is-dym-rollapp", rollerData.RollappID): true,
		"extra-codecs": []string{
			"ethermint",
		},
	}
	err := yamlconfig.UpdateNestedYAML(relayerConfigPath, updates)
	if err != nil {
		pterm.Error.Printf("Error updating YAML: %v\n", err)
		return err
	}

	return nil
}

func ValidateIbcPathChains(relayerHome, raID string, hd consts.HubData) (bool, error) {
	var srcChainOk bool
	var dstChainOk bool
	var defaultPathOk bool
	var relayerConfigExists bool
	var err error

	relayerConfigPath := GetConfigFilePath(relayerHome)

	pterm.Info.Println("checking configuration")
	// 2. config file exists
	relayerConfigExists, err = filesystem.DoesFileExist(relayerConfigPath)
	if err != nil {
		return false, err
	}

	if relayerConfigExists {
		// 2.1. path exist
		defaultPathOk, err = VerifyDefaultPath(relayerHome)
		if err != nil {
			pterm.Error.Printf(
				"failed to verify relayer path %s: %v\n",
				consts.DefaultRelayerPath,
				err,
			)
		}

		if defaultPathOk {
			// 2.2. isHubChainPresent
			srcChainOk, err = VerifyPathSrcChain(relayerHome, hd)
			if err != nil {
				pterm.Error.Printf(
					"failed to verify source chain in relayer path: %v\n",
					err,
				)
			}

			// 2.3. isRollappChainPresent
			dstChainOk, err = VerifyPathDstChain(relayerHome, raID)
			if err != nil {
				return false, err
			}
		}
	}

	if !relayerConfigExists || !srcChainOk || !dstChainOk {
		return false, nil
	}

	return true, nil
}
