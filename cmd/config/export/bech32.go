package export

import (
	"path/filepath"
	"strings"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

func getBech32Prefix(rlpCfg config.RollappConfig) (string, error) {
	rollappSeqAddrInfo, err := utils.GetAddressInfoBinary(
		utils.KeyConfig{
			Dir: filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp),
			ID:  consts.KeysIds.RollappSequencer,
		}, rlpCfg.RollappBinary,
	)
	if err != nil {
		return "", err
	}
	return strings.Split(rollappSeqAddrInfo.Address, "1")[0], nil
}
