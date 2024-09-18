package export

import (
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/utils/config"
)

func getBech32Prefix(rlpCfg config.RollappConfig) (string, error) {
	rollappSeqAddrInfo, err := utils.GetAddressInfoBinary(
		utils.KeyConfig{
			Dir:         consts.ConfigDirName.Rollapp,
			ID:          consts.KeysIds.RollappSequencer,
			ChainBinary: consts.Executables.RollappEVM,
			Type:        "",
		}, rlpCfg.Home,
	)
	if err != nil {
		return "", err
	}
	return strings.Split(rollappSeqAddrInfo.Address, "1")[0], nil
}
