package utils

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/roller"
)

func AddGlobalFlags(command *cobra.Command) {
	command.PersistentFlags().StringP(
		FlagNames.Home, "", roller.GetRootDir(), "The directory of the roller config files")
}

var FlagNames = struct {
	Home string
}{
	Home: "home",
}
