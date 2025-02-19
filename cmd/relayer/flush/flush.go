package flush

import (
	"github.com/pterm/pterm"
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

			spinner, _ := pterm.DefaultSpinner.Start("flushing relayer...")
			// nolint errcheck
			defer spinner.Stop()

			relayerutils.Flush(home)

			return nil
		},
	}

	return cmd
}
