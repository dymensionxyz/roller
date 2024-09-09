package get

import (
	"fmt"
	"os/exec"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve the current sequencer bond amount",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("getting sequencer bond")
			home := cmd.Flag(utils.FlagNames.Home).Value.String()

			rollerData, err := tomlconfig.LoadRollerConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			address, err := sequencer.GetHubSequencerAddress(rollerData)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer address", err)
				return
			}

			c := exec.Command(
				consts.Executables.Dymension,
				"q",
				"sequencer",
				"show-sequencer",
				address,
			)
			fmt.Println(c.String())

			out, err := bash.ExecCommandWithStdout(c)
			if err != nil {
				fmt.Println("failed to retrieve sequencer", err)
				return
			}

			fmt.Println(out.String())
		},
	}

	return cmd
}
