package init

import (
	"os"
	"os/exec"
	"path/filepath"
)

func initializeLightNodeConfig() error {
	celestia_executable_path := "/Users/itaylevy/go/bin/celestia"
	initLightNodeCmd := exec.Command(celestia_executable_path, "light", "init", "--p2p.network", "arabica", "--node.store", filepath.Join(os.Getenv("HOME"), configDirName.LightNode))
	err := initLightNodeCmd.Run()
	if err != nil {
		return err
	}
	return nil
}
