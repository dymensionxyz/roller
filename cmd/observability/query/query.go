package query

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/healthagent"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "query",
		Short: "Show the status of the sequencer on the local machine.",
		Run: func(cmd *cobra.Command, args []string) {
			metrics := []string{
				"dymint_mempool_size",
				"rollapp_pending_submissions_skew_batches",
				"rollapp_hub_height",
				"rollapp_consecutive_failed_da_submissions",
			}
			for _, metric := range metrics {
				value, err := healthagent.QueryPromMetric("localhost", "2112", metric)
				if err != nil {
					fmt.Printf("%s: failed to query metric: %s\n", metric, err)
				} else {
					fmt.Printf("%s: %d\n", metric, value)
				}
			}
		},
	}
	return cmd
}
