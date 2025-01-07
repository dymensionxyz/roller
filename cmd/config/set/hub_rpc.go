package set

import (
	"fmt"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

// setHubRPC function  
func setHubRPC(rlpCfg roller.RollappConfig, value string) error {
	rlpCfg.HubData.RpcUrl = value
	rlpCfg.HubData.ArchiveRpcUrl = value
	if err := roller.WriteConfig(rlpCfg); err != nil {
		return err
	}

	updates := map[string]interface{}{
		fmt.Sprintf("chains.%s.value.rpc-addr", rlpCfg.HubData.ID): value,
	}

	rlyConfigPath := filepath.Join(
		rlpCfg.Home,
		consts.ConfigDirName.Relayer,
		"config",
		"config.yaml",
	)
	err := yamlconfig.UpdateNestedYAML(rlyConfigPath, updates)
	if err != nil {
		pterm.Error.Printf("Error updating YAML: %v\n", err)
		return err
	}

	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	err = tomlconfig.UpdateFieldInFile(dymintTomlPath, "settlement_node_address", value)
	if err != nil {
		return err
	}

	return servicemanager.RestartSystemServices([]string{"rollapp", "relayer"}, rlpCfg.Home)
}
