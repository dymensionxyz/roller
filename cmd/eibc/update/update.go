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
		Short: "Update the Dymension eIBC client.",
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("preparing update")

			bvi, err := firebase.GetDependencyVersions()
			if err != nil {
				pterm.Error.Println("failed to fetch binary versions: ", err)
				return
			}

			eibcDep := dependencytypes.Dependency{
				DependencyName:  "eibc-client",
				RepositoryOwner: "dymensionxyz",
				RepositoryName:  "eibc-client",
				RepositoryUrl:   "https://github.com/dymensionxyz/eibc-client",
				Release:         bvi.EibcClient,
				Binaries: []types.BinaryPathPair{
					{
						Binary:            "eibc-client",
						BinaryDestination: consts.Executables.Eibc,
					},
				},
			}

			_ = servicemanager.StopSystemServices([]string{"eibc"})
			err = dependencies.InstallBinaryFromRelease(eibcDep)
			if err != nil {
				pterm.Error.Println("failed to install eibc client: ", err)
				return
			}

			_ = servicemanager.StartSystemServices([]string{"eibc"})
		},
	}

	return cmd
}
