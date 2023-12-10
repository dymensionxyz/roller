package migrate

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
)

type VersionMigratorV1000 struct{}

func (v *VersionMigratorV1000) ShouldMigrate(prevVersion VersionData) bool {
	return prevVersion.Major < 1
}

func (v *VersionMigratorV1000) PerformMigration(rlpCfg config.RollappConfig) error {
	fmt.Println("💈 Migrating Rollapp key...")
	migrateRollappKeyCmd := exec.Command(consts.Executables.RollappEVM, "keys", "migrate", "--home", rlpCfg.Home+"/relayer/keys/"+rlpCfg.RollappID, "--keyring-backend", "test")
	out, err := utils.ExecBashCommandWithStdout(migrateRollappKeyCmd)
	if err != nil {
		return err
	}
	fmt.Println(out.String())
	fmt.Println("💈 Migrating Hub key...")
	migrateHubKeyCmd := exec.Command(consts.Executables.Dymension, "keys", "migrate", "--home", rlpCfg.Home+"/relayer/keys/"+rlpCfg.HubData.ID, "--keyring-backend", "test")
	out, err = utils.ExecBashCommandWithStdout(migrateHubKeyCmd)
	if err != nil {
		return err
	}
	fmt.Println(out.String())
	fmt.Println("💈 Updating relayer configuration to match new relayer key...")
	if err := relayer.UpdateRlyConfigValue(rlpCfg, []string{"chains", rlpCfg.RollappID, "value", "extra-codecs"}, []string{"ethermint"}); err != nil {
		return err
	}
	return nil

}
