package initconfig

import (
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func initializeLightNodeConfig(initConfig utils.RollappConfig) error {
	initLightNodeCmd := exec.Command(consts.Executables.Celestia, "light", "init", "--p2p.network", consts.DefaultCelestiaNetwork, "--node.store", filepath.Join(initConfig.Home, consts.ConfigDirName.DALightNode))
	err := initLightNodeCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
