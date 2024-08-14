package initconfig

import (
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/genesis"
)

func InitializeRollappConfig(initConfig config.RollappConfig, hd consts.HubData) error {
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

	if initConfig.HubData.ID != "mock" {
		err := genesis.DownloadGenesis(home, initConfig)
		if err != nil {
			pterm.Error.Println("failed to download genesis file: ", err)
			return err
		}

		isChecksumValid, err := genesis.CompareGenesisChecksum(home, initConfig.RollappID, hd)
		if !isChecksumValid {
			return err
		}
	}

	_, err := bash.ExecCommandWithStdout(initRollappCmd)
	if err != nil {
		return err
	}

	err = setRollappConfig(initConfig)
	if err != nil {
		return err
	}

	return nil
}

func setRollappConfig(rlpCfg config.RollappConfig) error {
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
