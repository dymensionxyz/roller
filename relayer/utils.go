package relayer

import (
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func GetHomeDir(home string) string {
	return filepath.Join(home, consts.ConfigDirName.Relayer)
}

func GetConfigFilePath(relayerHome string) string {
	return filepath.Join(relayerHome, "config", "config.yaml")
}
