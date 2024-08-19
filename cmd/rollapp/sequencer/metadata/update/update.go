package update

import (
	"fmt"

	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update [metadata-file.json]",
		Short: "Update the sequencer metadata",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("update")
		},
	}

	return cmd
}
