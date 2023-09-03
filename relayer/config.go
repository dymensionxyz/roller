package relayer

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"os/exec"
	"path/filepath"
)

func CreatePath(rlpCfg config.RollappConfig) error {
	relayerHome := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer)
	setSettlementCmd := exec.Command(consts.Executables.Relayer, "chains", "set-settlement",
		rlpCfg.HubData.ID, "--home", relayerHome)
	if err := setSettlementCmd.Run(); err != nil {
		return err
	}
	newPathCmd := exec.Command(consts.Executables.Relayer, "paths", "new", rlpCfg.HubData.ID, rlpCfg.RollappID,
		consts.DefaultRelayerPath, "--home", relayerHome)
	if err := newPathCmd.Run(); err != nil {
		return err
	}
	return nil
}

func DeletePath(rlpCfg config.RollappConfig, pathName string) error {
	deletePathCmd := exec.Command(consts.Executables.Relayer, "paths", "delete", pathName, "--home",
		filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer))
	_, err := utils.ExecBashCommandWithStdout(deletePathCmd)
	if err != nil {
		return err
	}
	return nil
}

type ChainConfig struct {
	ID            string
	RPC           string
	Denom         string
	AddressPrefix string
	GasPrices     string
}
