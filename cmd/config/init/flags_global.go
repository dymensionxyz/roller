package initconfig

import (
	"os"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/roller"
)

func AddGlobalFlags(command *cobra.Command) {
	home, err := roller.GetRootDir()
	if err != nil {
		pterm.Error.Printf("failed to get roller root directory: %v", err)
		os.Exit(1)
	}
	command.PersistentFlags().StringP(
		GlobalFlagNames.Home, "", home, "The directory of the roller config files")
}

var GlobalFlagNames = struct {
	Home string
}{
	Home: "home",
}
