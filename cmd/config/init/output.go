package initconfig

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
)

func printInitOutput(addresses []utils.AddressData, rollappId string) {
	fmt.Printf("💈 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n", rollappId)
	requireFundingKeys := map[string]string{
		consts.KeysIds.HubSequencer: "Sequencer",
		consts.KeysIds.HubRelayer:   "Relayer",
		consts.KeysIds.DALightNode:  "Celestia",
	}
	filteredAddresses := make([]utils.AddressData, 0)
	for _, address := range addresses {
		if newName, ok := requireFundingKeys[address.Name]; ok {
			address.Name = newName
			filteredAddresses = append(filteredAddresses, address)
		}
	}
	utils.PrintAddresses(filteredAddresses)
	fmt.Printf("\n🔔 Please fund these addresses to register and run the rollapp.\n")
}
