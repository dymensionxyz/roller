package initconfig

import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/sequencer"
	"github.com/dymensionxyz/roller/utils/bash"
	genesisutils "github.com/dymensionxyz/roller/utils/genesis"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

func InitializeRollappConfig(
	initConfig *roller.RollappConfig,
	raResp rollapp.ShowRollappResponse,
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

	if initConfig.HubData.ID != "mock" {
		err := genesisutils.DownloadGenesis(initConfig.Home, raResp.Rollapp.Metadata.GenesisUrl)
		if err != nil {
			pterm.Error.Println("failed to download genesis file: ", err)
			return err
		}

		genesisFilePath := genesisutils.GetGenesisFilePath(initConfig.Home)
		err = genesisutils.VerifyGenesisChainID(genesisFilePath, initConfig.RollappID)
		if err != nil {
			return err
		}

		isChecksumValid, err := genesisutils.CompareGenesisChecksum(
			initConfig.Home,
			raResp.Rollapp.RollappId,
			initConfig.HubData,
		)
		if !isChecksumValid {
			return err
		}

		as, err := genesisutils.GetAppStateFromGenesisFile(initConfig.Home)
		if err != nil {
			return err
		}

		rollappBaseDenom := as.Staking.Params.BondDenom
		var rollappDenom string

		if strings.HasPrefix(as.Staking.Params.BondDenom, "ibc/") {
			rollappDenom = as.Staking.Params.BondDenom
		} else {
			rollappDenom = rollappBaseDenom[1:]
		}

		initConfig.BaseDenom = rollappBaseDenom
		initConfig.Denom = rollappDenom
	}

	err = setRollappConfig(*initConfig)
	if err != nil {
		return err
	}

	return nil
}

func setRollappConfig(rlpCfg roller.RollappConfig) error {
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
