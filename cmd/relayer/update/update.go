package update

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/dependencies/types"
	dependencytypes "github.com/dymensionxyz/roller/utils/dependencies/types"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/firebase"
	"github.com/dymensionxyz/roller/utils/roller"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update the Dymension's relayer version",
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("fetching environment from roller config")
			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Println("failed to expand home directory")
				return
			}
			localRollerConfig, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}
			pterm.Info.Println("environment:", localRollerConfig.HubData.Environment)

			pterm.Info.Println("preparing update")

			bvi, err := firebase.GetDependencyVersions(localRollerConfig.HubData.Environment)
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
