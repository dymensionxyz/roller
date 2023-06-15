package initconfig

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"os/exec"
	"path/filepath"
)

func initializeLightNodeConfig(initConfig InitConfig) error {
	initLightNodeCmd := exec.Command(consts.Executables.Celestia, "light", "init", "--p2p.network", "arabica", "--node.store", filepath.Join(initConfig.Home, consts.ConfigDirName.DALightNode))
	err := initLightNodeCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
