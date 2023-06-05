package version

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	BuildVersion = "v0.0.0"
	BuildTime    = "2023-06-05T14:28:30Z"
	BuildCommit  = "fac8ada6eef7d846971274efea1127ab00909b03"
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
