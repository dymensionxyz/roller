package migrate

import (
	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
)

type VersionMigratorV0118 struct{}

func (v *VersionMigratorV0118) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1 && prevVersion.Minor < 2 && prevVersion.Patch < 18
}

func (v *VersionMigratorV0118) PerformMigration(rlpCfg config.RollappConfig) error {
	rlyCfg, err := relayer.ReadRlyConfig(rlpCfg.Home)
	if err != nil {
		return err
	}
	hubRpcAddress := rlyCfg["chains"].(map[interface{}]interface{})[rlpCfg.HubData.ID].(map[interface{}]interface{})["value"].(map[interface{}]interface{})["rpc-addr"].(string)
	if hubRpcAddress == initconfig.Hubs[initconfig.FroopylandHubName].RPC_URL {
		if err := relayer.UpdateRlyConfigValue(rlpCfg, []string{"chains", rlpCfg.HubData.ID, "value", "rpc-addr"}, initconfig.Hubs[initconfig.FroopylandHubName].ARCHIVE_RPC_URL); err != nil {
			return err
		}

	}
	return nil

}
