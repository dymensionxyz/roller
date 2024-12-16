package update

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	dependencytypes "github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/firebase"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the Dymension's relayer version",
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("preparing update")

			bvi, err := firebase.GetDependencyVersions()
			if err != nil {
				pterm.Error.Println("failed to fetch binary versions: ", err)
				return
			}

			rlyDep := dependencytypes.Dependency{
				DependencyName:  "go-relayer",
				RepositoryOwner: "dymensionxyz",
				RepositoryName:  "go-relayer",
				RepositoryUrl:   "https://github.com/dymensionxyz/go-relayer",
				Release:         bvi.Relayer,
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "rly",
						BinaryDestination: consts.Executables.Relayer,
					},
				},
			}

			_ = servicemanager.StopSystemServices([]string{"relayer"})
			err = dependencies.InstallBinaryFromRelease(rlyDep)
			if err != nil {
				pterm.Error.Println("failed to install eibc client: ", err)
				return
			}

			_ = servicemanager.StartSystemServices([]string{"relayer"})
		},
	}

	return cmd
}
