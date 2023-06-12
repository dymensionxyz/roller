package utils

import (
	"os"
	"path/filepath"
)

func GetRollerRootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".roller")
}
