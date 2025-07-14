package filesystem

import (
	"os"
	"os/exec"

	"github.com/dymensionxyz/roller/utils/dependencies/types"
)

func IsAvailable(binary string) bool {
	_, err := exec.LookPath(binary)
	return err == nil
}

func BinariesExist(dep types.Dependency) bool {
	for _, bin := range dep.Binaries {
		if _, err := os.Stat(bin.BinaryDestination); err != nil {
			return false
		}
	}
	return true
}
