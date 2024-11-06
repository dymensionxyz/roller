package initconfig

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/roller"
)

func AddGlobalFlags(command *cobra.Command) {
	home := roller.GetRootDir()
	command.PersistentFlags().StringP(
		GlobalFlagNames.Home, "", home, "The directory of the roller config files")
}

var GlobalFlagNames = struct {
	Home string
}{
	Home: "home",
}
