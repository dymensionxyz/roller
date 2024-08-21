package initconfig

import (
	"encoding/json"
	"errors"
	"os/exec"
	"path/filepath"

	comettypes "github.com/cometbft/cometbft/types"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
)

func InitializeRollappConfig(initConfig *config.RollappConfig, hd consts.HubData) error {
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

	_, err := bash.ExecCommandWithStdout(initRollappCmd)
	if err != nil {
		return err
	}

	if initConfig.HubData.ID != "mock" {
		err := genesisutils.DownloadGenesis(initConfig.Home, *initConfig)
		if err != nil {
			pterm.Error.Println("failed to download genesis file: ", err)
			return err
		}

		isChecksumValid, err := genesisutils.CompareGenesisChecksum(
			initConfig.Home,
			initConfig.RollappID,
			hd,
		)
		if !isChecksumValid {
			return err
		}

		genesisDoc, err := comettypes.GenesisDocFromFile(
			genesisutils.GetGenesisFilePath(initConfig.Home),
		)
		if err != nil {
			return err
		}

		// TODO: refactor
		var need genesisutils.AppState
		j, err := genesisDoc.AppState.MarshalJSON()
		if err != nil {
			return err
		}

		err = json.Unmarshal(j, &need)
		if err != nil {
			return err
		}

		if len(need.Bank.Supply) == 0 {
			return errors.New("token supply is not defined in the genesis file")
		}

		rollappBaseDenom := need.Bank.Supply[0].Denom
		rollappDenom := rollappBaseDenom[1:]

		initConfig.BaseDenom = rollappBaseDenom
		initConfig.Denom = rollappDenom
	}

	err = setRollappConfig(*initConfig)
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
