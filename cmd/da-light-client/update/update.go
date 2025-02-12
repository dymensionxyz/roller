package update

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/dependencies"
	servicemanager "github.com/dymensionxyz/roller/utils/service_manager"
)

const (
	rpcEndpointFlag     = "rpc-endpoint"
	metricsEndpointFlag = "metrics-endpoint"
)

func Cmd() *cobra.Command {
	runCmd := &cobra.Command{
		Use:   "update",
		Short: "Runs the DA light client.",
		Run: func(cmd *cobra.Command, args []string) {
			pterm.Info.Println("stopping existing system services, if any...")
			err := servicemanager.StopSystemServices([]string{"da-light-client"})
			if err != nil {
				pterm.Error.Println("failed to stop system services: ", err)
				return
			}

			dep := dependencies.DefaultCelestiaNodeDependency()
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
