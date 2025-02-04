package rngoracle

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/services"
	loadservices "github.com/dymensionxyz/roller/cmd/services/load"
	restartservices "github.com/dymensionxyz/roller/cmd/services/restart"
	startservices "github.com/dymensionxyz/roller/cmd/services/start"
	stopservices "github.com/dymensionxyz/roller/cmd/services/stop"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rng",
		Short: "Commands related to Rng Oracle smart contract deployment and client operation",
	}

	cmd.AddCommand(DeployCmd())
	cmd.AddCommand(StartCmd())

	sl := []string{"rng"}
	cmd.AddCommand(
		services.Cmd(
			loadservices.Cmd(sl, "rng"),
			startservices.OracleCmd("rng"),
			restartservices.Cmd(sl),
			stopservices.Cmd(sl),
			// logservices.RollappCmd(),
		),
	)

	return cmd
}
