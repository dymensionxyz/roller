package initconfig

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"strings"
)

func printInitOutput(rollappConfig utils.RollappConfig, addresses []utils.AddressData, rollappId string) {
	fmt.Printf("💈 RollApp '%s' configuration files have been successfully generated on your local machine. Congratulations!\n\n", rollappId)
	fmt.Println(FormatTokenSupplyLine(rollappConfig))
	fmt.Println()
	utils.PrintAddresses(formatAddresses(addresses))
	fmt.Printf("\n🔔 Please fund these addresses to register and run the rollapp.\n")
}

func formatAddresses(addresses []utils.AddressData) []utils.AddressData {
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
	return filteredAddresses
}

func FormatTokenSupplyLine(rollappConfig utils.RollappConfig) string {
	displayDenom := strings.ToUpper(rollappConfig.Denom[1:])
	return fmt.Sprintf("💰 Total Token Supply: %s %s. Note that 1 %s == 1 * 10^%d %s (like 1 ETH == 1 * 10^18 wei).",
		addCommasToNum(rollappConfig.TokenSupply), displayDenom, displayDenom, rollappConfig.Decimals, "u"+displayDenom)
}

func addCommasToNum(number string) string {
	var result strings.Builder
	n := len(number)

	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteByte(number[i])
	}
	return result.String()
}
