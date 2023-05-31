package init

import (
	"os"
	"path/filepath"
)

func getRollerRootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".roller")
}
