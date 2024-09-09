package get

import (
	"encoding/json"
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
				"--output",
				"json",
				"--node", rollerData.HubData.RPC_URL,
				"--chain-id", rollerData.HubData.ID,
			)

			var GetSequencerResponse sequencer.ShowSequencerResponse
			out, err := bash.ExecCommandWithStdout(c)
			if err != nil {
				fmt.Println("failed to retrieve sequencer", err)
				return
			}
			err = json.Unmarshal(out.Bytes(), &GetSequencerResponse)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer", err)
			}
			pterm.DefaultSection.WithIndentCharacter("💈").
				Printf("%s bonded tokens", address)
			fmt.Println(GetSequencerResponse.Sequencer.Tokens.String())
		},
	}

	return cmd
}
