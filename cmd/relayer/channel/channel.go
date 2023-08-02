package channel

import "github.com/spf13/cobra"

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channel",
		Short: "Commands for managing the relayer channel",
	}
	return cmd
}
