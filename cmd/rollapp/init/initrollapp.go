package initrollapp

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/roller"
)

func InitializeRollappNode(
	initConfig roller.RollappConfig,
) error {
	raHomeDir := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)

	initRollappCmd := exec.Command(
		initConfig.RollappBinary,
		"init",
		consts.KeysIds.HubSequencer,
		"--chain-id",
		initConfig.RollappID,
		"--home",
		raHomeDir,
	)

	_, err := bash.ExecCommandWithStdout(initRollappCmd)
	if err != nil {
		return err
	}

	err = writeConfigFilesForRollapp(initConfig)
	if err != nil {
		return err
	}

	return nil
}

func writeConfigFilesForRollapp(rlpCfg roller.RollappConfig) error {
	if err := sequencer.SetAppConfig(rlpCfg); err != nil {
		return err
	}
	if err := sequencer.SetTMConfig(rlpCfg); err != nil {
		return err
	}
	if err := sequencer.SetDefaultDymintConfig(rlpCfg); err != nil {
		return err
	}
	return nil
}
