package oracle

import (
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/oracle/deploy"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oracle",
		Short: "Commands related to RollApp's component observability",
	}

	cmd.AddCommand(deploy.Cmd())

	return cmd
}
