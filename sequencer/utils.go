package sequencer

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"path/filepath"
)

func GetDymintFilePath(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config", "dymint.toml")
}
