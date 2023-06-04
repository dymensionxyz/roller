package init

import (
	"os/exec"
	"path/filepath"
)

func initializeLightNodeConfig() error {
	initLightNodeCmd := exec.Command(celestiaExecutablePath, "light", "init", "--p2p.network", "arabica", "--node.store", filepath.Join(getRollerRootDir(), configDirName.DALightNode))
	err := initLightNodeCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
