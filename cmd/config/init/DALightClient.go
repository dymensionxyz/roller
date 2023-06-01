package init

import (
	"os/exec"
	"path/filepath"
)

func initializeLightNodeConfig(initConfig InitConfig) error {
	initLightNodeCmd := exec.Command(celestiaExecutablePath, "light", "init", "--p2p.network", "arabica", "--node.store", filepath.Join(initConfig.Home, configDirName.DALightNode))
	err := initLightNodeCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
