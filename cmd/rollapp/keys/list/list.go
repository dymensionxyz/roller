package list

import (
	"encoding/json"
	"fmt"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all rollapp addresses.",
		Run: func(cmd *cobra.Command, args []string) {
			home := cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String()
			rollerData, err := roller.LoadConfig(home)
			errorhandling.PrettifyErrorIfExists(err)
			addresses := make([]keys.KeyInfo, 0)

			var kc []keys.KeyConfig
			if rollerData.HubData.ID != "mock" {
				kc = keys.GetSequencerKeysConfig(rollerData.KeyringBackend)
			} else {
				kc = keys.GetMockSequencerKeyConfig(rollerData)
			}

			ki, err := kc[0].Info(home)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer info: ", err)
				return
			}
			seqPubKey, err := keys.GetSequencerPubKey(rollerData)
			if err != nil {
				pterm.Error.Println("failed to retrieve sequencer public key: ", err)
				return
			}
			ki.PubKey = seqPubKey

			addresses = append(addresses, *ki)

			for _, address := range addresses {
				address.Print(keys.WithName(), keys.WithPubKey())
			}
		},
	}
	// cmd.Flags().StringP(flagNames.outputType, "", "text", "Output format (text|json)")
	return cmd
}

// nolint: unused
func printAsJSON(addresses []keys.KeyInfo) error {
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
