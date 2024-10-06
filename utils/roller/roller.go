package roller

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"
)

func GetRootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".roller")
}

func GetConfigPath(home string) string {
	return filepath.Join(home, "roller.toml")
}

func CreateConfigFile(home string) (bool, error) {
	rollerConfigFilePath := GetConfigPath(home)

	_, err := os.Stat(rollerConfigFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			pterm.Info.Println("roller.toml not found, creating")
			_, err := os.Create(rollerConfigFilePath)
			if err != nil {
				pterm.Error.Printf(
					"failed to create %s: %v", rollerConfigFilePath, err,
				)
				return false, err
			}

			return true, nil
		}
	}

	return false, nil
}
