package migrate

import (
	"fmt"

	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/data_layer/celestia"
)

type VersionMigratorV1005 struct{}

func (v *VersionMigratorV1005) ShouldMigrate(prevVersion VersionData) bool {
	fmt.Println(prevVersion)
	if prevVersion.Major < 1 || (prevVersion.Major == 1 && prevVersion.Minor < 1 && prevVersion.Patch < 5) {
		return true
	}
	return false
}

func (v *VersionMigratorV1005) PerformMigration(rlpCfg config.RollappConfig) error {
	// Check if celestia is the data layer.
	if rlpCfg.DA != config.Celestia {
		return nil
	}
	// If it is celestia, re-init the light node and ask the user to fund the address
	celestiaClient := celestia.NewCelestia(rlpCfg.Home)
	celestiaClient.InitializeLightNodeConfig()
	address, err := celestiaClient.GetDAAccountAddress()
	if err != nil {
		return err
	}
	fmt.Println("💈 Celestia client migrated successfully!")
	fmt.Println("🔔 Please fund the following from the celestia-faucet on discord 🔔", ":", address)
	return nil

}
