package export

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"path/filepath"
	"strings"
)

func getBech32Prefix(rlpCfg config.RollappConfig) (string, error) {
	rollappSeqAddr, err := utils.GetAddressBinary(utils.KeyConfig{
		Dir: filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp),
		ID:  consts.KeysIds.RollappSequencer,
	}, rlpCfg.RollappBinary)
	if err != nil {
		return "", err
	}
	return strings.Split(rollappSeqAddr, "1")[0], nil
}
