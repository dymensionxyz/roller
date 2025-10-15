package update

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/firebase"
	"github.com/dymensionxyz/roller/utils/roller"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

const (
	rpcEndpointFlag     = "rpc-endpoint"
	metricsEndpointFlag = "metrics-endpoint"
)

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "update",
		Short: "Updates the DA light client binary.",
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

			pterm.Info.Println("stopping existing system services, if any...")
			err = servicemanager.StopSystemServices([]string{"da-light-client"})
			if err != nil {
				pterm.Error.Println("failed to stop system services: ", err)
				return
			}

			bvi, err := firebase.GetDependencyVersions(localRollerConfig.HubData.Environment)
			if err != nil {
				pterm.Error.Println("failed to get dependency versions: ", err)
				return
			}

			dep := dependencies.CelestiaNodeDependency(*bvi)
			err = dependencies.InstallBinaryFromRepo(
				dep, dep.DependencyName,
			)
			if err != nil {
				pterm.Error.Println("failed to install binary: ", err)
				return
			}

			pterm.Info.Println("stopping existing system services, if any...")
			err = servicemanager.StartSystemServices([]string{"da-light-client"})
			if err != nil {
				pterm.Error.Println("failed to stop system services: ", err)
				return
			}
		},
	}

	addFlags(runCmd)
	return runCmd
}

func addFlags(cmd *cobra.Command) {
	cmd.Flags().
		StringP(rpcEndpointFlag, "", "rpc-mocha.pops.one", "The DA rpc endpoint to connect to.")
	cmd.Flags().
		StringP(metricsEndpointFlag, "", "", "The OTEL collector metrics endpoint to connect to.")
}
