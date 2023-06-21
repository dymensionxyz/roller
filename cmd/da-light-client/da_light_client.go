package da_light_client

import (
	da_start "github.com/dymensionxyz/roller/cmd/da-light-client/start"
	"github.com/spf13/cobra"
)

func DALightClientCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "da-light-client",
		Short: "Commands for running and managing the data availability light client.",
	}
	cmd.AddCommand(da_start.Cmd())
	return cmd
}
