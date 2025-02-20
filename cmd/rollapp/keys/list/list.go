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

			aki, err := keys.All(rollerData, rollerData.HubData)
			if err != nil {
				pterm.Error.Println("failed to get all keys", err)
				return
			}

			for _, addrData := range aki {
				addrData.Print(keys.WithName())
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
