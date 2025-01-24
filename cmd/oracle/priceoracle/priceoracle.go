package priceoracle

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
		Use:   "price-oracle",
		Short: "Commands related to Price Oracle smart contract deployment and client operation",
	}

	cmd.AddCommand(DeployCmd())
	cmd.AddCommand(StartCmd())

	sl := []string{"oracle"}
	cmd.AddCommand(
		services.Cmd(
			loadservices.Cmd(sl, "oracle"),
			startservices.OracleCmd(),
			restartservices.Cmd(sl),
			stopservices.Cmd(sl),
			// logservices.RollappCmd(),
		),
	)

	return cmd
}
