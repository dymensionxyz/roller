package get

import (
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/sequencer"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "Retrieve the current sequencer bond amount",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Println("failed to load roller config file", err)
				return
			}

			address, err := sequencer.GetHubSequencerAddress(rollerData)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer address", err)
				return
			}

			bond, err := sequencer.GetSequencerBond(address, rollerData.HubData)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer bond", err)
				return
			}

			pterm.DefaultSection.WithIndentCharacter("💈").
				Printf("%s bonded tokens", address)
			fmt.Println(bond.String())
		},
	}

	return cmd
}
