package initconfig

import (
	"fmt"
	"strings"

	"github.com/dymensionxyz/roller/utils/config"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	datalayer "github.com/dymensionxyz/roller/data_layer"
)

func formatAddresses(
	rollappConfig config.RollappConfig,
	addresses []utils.KeyInfo,
) []utils.KeyInfo {
	damanager := datalayer.NewDAManager(rollappConfig.DA, rollappConfig.Home)
	requireFundingKeys := map[string]string{
		consts.KeysIds.HubSequencer: fmt.Sprintf("Sequencer, %s Hub", rollappConfig.HubData.ID),
		consts.KeysIds.HubRelayer:   fmt.Sprintf("Relayer, %s Hub", rollappConfig.HubData.ID),
		damanager.GetKeyName():      fmt.Sprintf("DA, %s Network", damanager.GetNetworkName()),
	}
	filteredAddresses := make([]utils.KeyInfo, 0)
	for _, address := range addresses {
		if newName, ok := requireFundingKeys[address.Name]; ok {
			address.Name = newName
			filteredAddresses = append(filteredAddresses, address)
		}
	}
	return filteredAddresses
}

func PrintTokenSupplyLine(rollappConfig config.RollappConfig) {
	pterm.DefaultSection.WithIndentCharacter("💰").Printf(
		"Total Token Supply: %s %s.",
		addCommasToNum(
			consts.DefaultTokenSupply[:len(consts.DefaultTokenSupply)-int(rollappConfig.Decimals)],
		),
		rollappConfig.Denom,
	)

	pterm.DefaultBasicText.Printf(
		"Note that 1 %s == 1 * 10^%d %s (like 1 ETH == 1 * 10^18 wei).\nThe total supply in base denom (%s) is "+pterm.Yellow(
			"%s%s",
		),
		rollappConfig.Denom,
		rollappConfig.Decimals,
		rollappConfig.BaseDenom,
		rollappConfig.BaseDenom,
		consts.DefaultTokenSupply,
		rollappConfig.BaseDenom,
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
