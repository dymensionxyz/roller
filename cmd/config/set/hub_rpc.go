package set

import (
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
)

func setHubRPC(rlpCfg config.RollappConfig, value string) error {
	rlpCfg.HubData.RPC_URL = value
	if err := config.WriteConfigToTOML(rlpCfg); err != nil {
		return err
	}
	if err := relayer.UpdateRlyConfigValue(rlpCfg, []string{"chains", rlpCfg.HubData.ID, "value", "rpc-addr"},
		value); err != nil {
		return err
	}
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	return utils.UpdateFieldInToml(dymintTomlPath, "node_address", value)
}
