package set

import (
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

func setHubRPC(rlpCfg roller.RollappConfig, value string) error {
	rlpCfg.HubData.RPC_URL = value
	if err := tomlconfig.Write(rlpCfg); err != nil {
		return err
	}
	if err := relayer.UpdateRlyConfigValue(
		rlpCfg, []string{"chains", rlpCfg.HubData.ID, "value", "rpc-addr"},
		value,
	); err != nil {
		return err
	}
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	return tomlconfig.UpdateFieldInToml(dymintTomlPath, "settlement_node_address", value)
}
