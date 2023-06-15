package utils

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func AddGlobalFlags(command *cobra.Command) {
	command.Flags().StringP(FlagNames.Home, "", GetRollerRootDir(), "The directory of the roller config files")
}

var FlagNames = struct {
	Home string
}{
	Home: "home",
}

func GetRollerRootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".roller")
}
