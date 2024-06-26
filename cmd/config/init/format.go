package initconfig

import (
	"fmt"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	datalayer "github.com/dymensionxyz/roller/data_layer"
)

func formatAddresses(
	rollappConfig config.RollappConfig,
	addresses []utils.AddressData,
) []utils.AddressData {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	requireFundingKeys := map[string]string{
		consts.KeysIds.HubSequencer: fmt.Sprintf("Sequencer, %s Hub", rollappConfig.HubData.ID),
		consts.KeysIds.HubRelayer:   fmt.Sprintf("Relayer, %s Hub", rollappConfig.HubData.ID),
		damanager.GetKeyName():      fmt.Sprintf("DA, %s Network", damanager.GetNetworkName()),
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

func FormatTokenSupplyLine(rollappConfig config.RollappConfig) string {
	displayDenom := strings.ToUpper(rollappConfig.Denom[1:])
	return fmt.Sprintf(
		"💰 Total Token Supply: %s %s. Note that 1 %s == 1 * 10^%d %s (like 1 ETH == 1 * 10^18 wei).",
		addCommasToNum(
			rollappConfig.TokenSupply,
		),
		displayDenom,
		displayDenom,
		rollappConfig.Decimals,
		"u"+displayDenom,
	)
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
