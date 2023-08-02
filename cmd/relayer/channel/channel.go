package channel

import (
	"github.com/dymensionxyz/roller/cmd/relayer/channel/show"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Commands for managing the relayer channel",
	}
	cmd.AddCommand(show.Cmd())
	return cmd
}
