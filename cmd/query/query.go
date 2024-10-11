package query

import (
	"fmt"

	"github.com/dymensionxyz/roller/utils/healthagent"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			submissions, err := healthagent.QueryFailedDaSubmissions("localhost", "2112")
			if err != nil {
				pterm.Error.Println(err)
				return
			}

			fmt.Println("sub", submissions)
		},
	}
	return cmd
}
