package initconfig

import (
	"fmt"

	"github.com/dymensionxyz/roller/cmd/consts"
	datalayer "github.com/dymensionxyz/roller/data_layer"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

func FormatAddresses(
	raCfg roller.RollappConfig,
	addresses []keys.KeyInfo,
) []keys.KeyInfo {
	damanager := datalayer.NewDAManager(raCfg.DA.Backend, raCfg.Home, raCfg.KeyringBackend)
	requireFundingKeys := map[string]string{
		consts.KeysIds.HubSequencer: fmt.Sprintf("Sequencer, %s Hub", raCfg.HubData.ID),
		consts.KeysIds.HubRelayer:   fmt.Sprintf("Relayer, %s Hub", raCfg.HubData.ID),
		damanager.GetKeyName():      fmt.Sprintf("DA, %s Network", damanager.GetNetworkName()),
	}
	filteredAddresses := make([]keys.KeyInfo, 0)
	for _, address := range addresses {
		if newName, ok := requireFundingKeys[address.Name]; ok {
			address.Name = newName
			filteredAddresses = append(filteredAddresses, address)
		}
	}
	return filteredAddresses
}
