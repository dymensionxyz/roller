package initconfig

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

func InitializeRollappConfig(initConfig config.RollappConfig) error {
	home := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)

	initRollappCmd := exec.Command(
		initConfig.RollappBinary,
		"init",
		consts.KeysIds.HubSequencer,
		"--chain-id",
		initConfig.RollappID,
		"--home",
		home,
	)

	_, err := utils.ExecBashCommandWithStdout(initRollappCmd)
	if err != nil {
		return err
	}

	return nil
}

// func setRollappConfig(rlpCfg config.RollappConfig) error {
// 	if err := sequencer.SetAppConfig(rlpCfg); err != nil {
// 		return err
// 	}
// 	if err := sequencer.SetTMConfig(rlpCfg); err != nil {
// 		return err
// 	}
// 	if err := sequencer.SetDefaultDymintConfig(rlpCfg); err != nil {
// 		return err
// 	}
// 	return nil
// }

func RollappConfigDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config")
}
