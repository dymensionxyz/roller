package sequencer

import (
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
)

func GetDymintFilePath(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config", "dymint.toml")
}
