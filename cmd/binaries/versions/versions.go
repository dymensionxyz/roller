package versions

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies"
)

func Cmd() *cobra.Command {
	versionCmd := &cobra.Command{
		Use:   "versions",
		Short: "Print all binary versions",
		Run: func(cmd *cobra.Command, args []string) {
			dymdVersion, err := dependencies.GetCurrentCommit(consts.Executables.Dymension)
			if err != nil {
				fmt.Println("failed to get dymd version:", err)
				return
			}
			fmt.Println("dymd version:", dymdVersion)
			celestiaVersion, err := dependencies.GetCurrentCommit(consts.Executables.Celestia)
			if err != nil {
				fmt.Println("failed to get celestia version:", err)
				return
			}
			fmt.Println("celestia version:", celestiaVersion)

			relayerVersion, err := dependencies.GetCurrentCommit(consts.Executables.Relayer)
			if err != nil {
				fmt.Println("failed to get relayer version:", err)
				return
			}
			fmt.Println("relayer version:", relayerVersion)

			eibcVersion, err := dependencies.GetCurrentCommit(consts.Executables.Eibc)
			if err != nil {
				fmt.Println("failed to get eibc version:", err)
				return
			}
			fmt.Println("eibc version:", eibcVersion)
		},
	}
	return versionCmd
}
