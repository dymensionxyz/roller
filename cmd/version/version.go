package version

import (
	"fmt"

	versionutils "github.com/dymensionxyz/roller/version"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of roller",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("💈 Roller version", versionutils.BuildVersion)
			fmt.Println("💈 Build time:", versionutils.BuildTime)
			fmt.Println("💈 Git commit:", versionutils.BuildCommit)
		},
	}
	return versionCmd
}
