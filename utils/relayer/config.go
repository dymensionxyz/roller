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
		): rollerData.HubData.ApiUrl,
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

type IbcPathChains struct {
	SrcChainOk          bool
	DstChainOk          bool
	DefaultPathOk       bool
	RelayerConfigExists bool
}

func ValidateIbcPathChains(relayerHome, raID string, hd consts.HubData) (*IbcPathChains, error) {
	var err error
	ibcPathChains := IbcPathChains{}

	relayerConfigPath := GetConfigFilePath(relayerHome)

	pterm.Info.Println("checking configuration")
	// 2. config file exists
	relayerConfigExists, err := filesystem.DoesFileExist(relayerConfigPath)
	if err != nil {
		return nil, err
	}
	ibcPathChains.RelayerConfigExists = relayerConfigExists

	if relayerConfigExists {
		// 2.1. path exist
		defaultPathOk, err := VerifyDefaultPath(relayerHome)
		if err != nil {
			pterm.Error.Printf(
				"failed to verify relayer path %s: %v\n",
				consts.DefaultRelayerPath,
				err,
			)
		}
		ibcPathChains.DefaultPathOk = defaultPathOk

		if defaultPathOk {
			// 2.2. isHubChainPresent
			srcChainOk, err := VerifyPathSrcChain(relayerHome, hd)
			if err != nil {
				pterm.Error.Printf(
					"failed to verify source chain in relayer path: %v\n",
					err,
				)
			}
			ibcPathChains.SrcChainOk = srcChainOk

			// 2.3. isRollappChainPresent
			dstChainOk, err := VerifyPathDstChain(relayerHome, raID)
			if err != nil {
				return &ibcPathChains, err
			}
			ibcPathChains.DstChainOk = dstChainOk
		} else {
			pterm.Error.Println("default path not found in relayer config")
		}
	}

	return &ibcPathChains, nil
}
