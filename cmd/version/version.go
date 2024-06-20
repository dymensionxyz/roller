package version

import (
	"fmt"

	"github.com/dymensionxyz/roller/version"

	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of roller",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("💈 Roller version", version.BuildVersion)
			fmt.Println("💈 Build time:", version.BuildTime)
			fmt.Println("💈 Git commit:", version.BuildCommit)
		},
	}
	return versionCmd
}
