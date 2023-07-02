package utils

import (
	"github.com/spf13/viper"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

func AddGlobalFlags(command *cobra.Command) {
	command.PersistentFlags().StringP(
		FlagNames.Home, "", GetRollerRootDir(), "The directory of the roller config files")
	err := viper.BindPFlag(FlagNames.Home, command.PersistentFlags().Lookup(FlagNames.Home))
	PrettifyErrorIfExists(err)
}

var FlagNames = struct {
	Home string
}{
	Home: "home",
}

func GetRollerRootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".roller")
}
