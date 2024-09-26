package funds

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/utils/bash"
	eibcutils "github.com/dymensionxyz/roller/utils/eibc"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "funds",
		Short: "Get an overview of available and pending fund status",
		Run: func(cmd *cobra.Command, args []string) {
			spin, _ := pterm.DefaultSpinner.Start("Fetching funds...")
			c := eibcutils.GetFundsCmd()
			out, err := bash.ExecCommandWithStdout(c)
			if err != nil {
				spin.Fail("failed to retrieve funds")
				pterm.Error.Println(err)
				return
			}
			spin.Success("Funds retrieved successfully")

			pterm.Info.Println("current eibc wallet fund distribution:")
			fmt.Println(out.String())
		},
	}
	return cmd
}
