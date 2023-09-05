package sequencer

import (
	"github.com/dymensionxyz/roller/cmd/consts"
	"math/big"
	"path/filepath"
)

func GetDymintFilePath(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config", "dymint.toml")
}

func NewStrBigIntWithoutErr(str string) *big.Int {
	bigInt := new(big.Int)
	bigInt.SetString(str, 10)
	return bigInt
}

var MinSufficientBalance = map[string]*big.Int{
	consts.Hubs[consts.LocalHubName].ID:      big.NewInt(1),
	consts.Hubs[consts.StagingHubName].ID:    big.NewInt(1),
	consts.Hubs[consts.FroopylandHubName].ID: NewStrBigIntWithoutErr("25000000000000000000"),
}
