package relayer

import (
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
)

func UpdateConfigWithDefaultValues(relayerHome string, rollerData roller.RollappConfig) error {
	pterm.Info.Println("updating application relayer config")
	relayerConfigPath := filepath.Join(relayerHome, "config", "config.yaml")
	updates := map[string]interface{}{
		fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.HubData.ID): 1.5,
		fmt.Sprintf("chains.%s.value.gas-prices", rollerData.HubData.ID): fmt.Sprintf(
			"20000000000%s",
			consts.Denoms.Hub,
		),
		fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.RollappID): 1.3,
		fmt.Sprintf("chains.%s.value.is-dym-hub", rollerData.HubData.ID):    true,
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
