package migrate

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

type VersionMigratorV1000 struct{}

func (v *VersionMigratorV1000) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1
}

func (v *VersionMigratorV1000) PerformMigration(rlpCfg config.RollappConfig) error {
	// Get relayer address in order to create a keys migration
	fmt.Println("💈 Migrating relayer keys...")
	_, err := utils.GetRelayerAddress(rlpCfg.Home, rlpCfg.HubData.ID)
	if err != nil {
		return err
	}
	_, err = utils.GetRelayerAddress(rlpCfg.Home, rlpCfg.RollappID)
	if err != nil {
		return err
	}
	return nil

}
