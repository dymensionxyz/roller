package initconfig

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/roller"
)

func AddGlobalFlags(command *cobra.Command) {
	command.PersistentFlags().StringP(
		GlobalFlagNames.Home, "", roller.GetRootDir(), "The directory of the roller config files")
}

var GlobalFlagNames = struct {
	Home string
}{
	Home: "home",
}
