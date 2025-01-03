package set

import (
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func setHubRPC(rlpCfg roller.RollappConfig, value string) error {
	rlpCfg.HubData.RpcUrl = value
	if err := roller.WriteConfig(rlpCfg); err != nil {
		return err
	}
	if err := relayer.UpdateRlyConfigValue(
		rlpCfg, []string{"chains", rlpCfg.HubData.ID, "value", "rpc-addr"},
		value,
	); err != nil {
		return err
	}
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	err := tomlconfig.UpdateFieldInFile(dymintTomlPath, "settlement_node_address", value)
	if err != nil {
		return err
	}

	return servicemanager.RestartSystemServices([]string{"rollapp", "relayer"}, rlpCfg.Home)
}
