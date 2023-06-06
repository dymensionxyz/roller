package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BuildVersion = "<version>"
	BuildTime    = "<build-time>"
	BuildCommit  = "<build-commit>"
)

func VersionCmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Print the version of roller",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("💈 Roller version", BuildVersion)
			fmt.Println("💈 Build time:", BuildTime)
			fmt.Println("💈 Git commit:", BuildCommit)
		},
	}
	return versionCmd
}
