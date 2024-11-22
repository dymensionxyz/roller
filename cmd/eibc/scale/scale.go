package scale

import (
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/bash"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scale [count]",
		Short: "Scale the number of fulfiller addresses to [count]",
		Long: `Scale the number of fulfiller wallets to [count]

fulfiller wallets are created to fulfill orders on behalf of the whale account

a good number to start with is 30 (default when initializing the eibc client)
`,
		Args: cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			count := args[0]

			c := eibcutils.GetScaleCmd(count)

			_, err := bash.ExecCommandWithStdout(c)
			if err != nil {
				pterm.Error.Println("failed to scale the number of fulfillers: ", err)
				return
			}
		},
	}
	return cmd
}
