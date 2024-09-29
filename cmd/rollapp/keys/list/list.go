package list

import (
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/errorhandling"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all rollapp addresses.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(utils.FlagNames.Home).Value.String()
			rollerData, err := tomlconfig.LoadRollerConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			addresses := make([]utils.KeyInfo, 0)

			var kc []utils.KeyConfig
			if rollerData.HubData.ID != "mock" {
				kc = initconfig.GetSequencerKeysConfig()
			} else {
				kc = initconfig.GetMockSequencerKeyConfig(rollerData)
			}

			ki, err := utils.GetAddressInfoBinary(kc[0], home)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer info: ", err)
				return
			}
			seqPubKey, err := utils.GetSequencerPubKey(rollerData)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer public key: ", err)
				return
			}
			ki.PubKey = seqPubKey

			addresses = append(addresses, *ki)

			for _, address := range addresses {
				address.Print(utils.WithName(), utils.WithPubKey())
			}
		},
	}
	// cmd.Flags().StringP(flagNames.outputType, "", "text", "Output format (text|json)")
	return cmd
}

// nolint: unused
func printAsJSON(addresses []utils.KeyInfo) error {
	addrMap := make(map[string]string)
	for _, addrData := range addresses {
		addrMap[addrData.Name] = addrData.Address
	}
	data, err := json.MarshalIndent(addrMap, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshalling data %s", err)
	}
	fmt.Println(string(data))
	return nil
}
