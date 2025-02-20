package flush

import (
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	relayerutils "github.com/dymensionxyz/roller/utils/relayer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flush",
		Short: "Flush the relayer",
		RunE: func(cmd *cobra.Command, args []string) error {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			relayerutils.Flush(home)

			return nil
		},
	}

	return cmd
}
