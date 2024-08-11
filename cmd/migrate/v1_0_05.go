package migrate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config"
)

type VersionMigratorV1005 struct{}

func (v *VersionMigratorV1005) ShouldMigrate(prevVersion VersionData) bool {
	if prevVersion.Major < 1 ||
		(prevVersion.Major == 1 && prevVersion.Minor < 1 && prevVersion.Patch < 5) {
		return true
	}
	return false
}

func (v *VersionMigratorV1005) PerformMigration(rlpCfg config.RollappConfig) error {
	// If the DA is not celestia, no-op
	if rlpCfg.DA != consts.Celestia {
		return nil
	}
	// Update dymint config with celestia new config
	dymintTomlPath := sequencer.GetDymintFilePath(rlpCfg.Home)
	da := datalayer.NewDAManager(rlpCfg.DA, rlpCfg.Home)
	sequencerDaConfig := da.GetSequencerDAConfig(consts.NodeType.Sequencer)
	if sequencerDaConfig == "" {
		return nil
	}
	if err := utils.UpdateFieldInToml(dymintTomlPath, "da_config", sequencerDaConfig); err != nil {
		return err
	}
	// Delete previous celestia data directory
	celestiaDataDir := filepath.Join(rlpCfg.Home, consts.ConfigDirName.DALightNode, "data")
	// Delete the celestia DataDir
	if err := os.RemoveAll(celestiaDataDir); err != nil {
		return err
	}
	// re-init the light node and ask the user to fund the address
	celestiaClient := celestia.NewCelestia(rlpCfg.Home)
	_, err := celestiaClient.InitializeLightNodeConfig()
	if err != nil {
		return err
	}
	address, err := celestiaClient.GetDAAccountAddress()
	if err != nil {
		return err
	}
	fmt.Println("💈 Celestia client migrated successfully!")
	fmt.Println("🔔 Please fund the following from the celestia-faucet on discord 🔔", ":", address)
	return nil
}
