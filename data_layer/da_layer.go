package datalayer

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/data_layer/celestia"
)

type DataLayer interface {
	GetDAAccountAddress() (string, error)
	InitializeLightNodeConfig() error
	CheckDABalance() ([]utils.NotFundedAddressData, error)
	GetStartDACmd(rpcEndpoint string) *exec.Cmd
	GetDAAccData(c config.RollappConfig) (*utils.AccountData, error)

	GetLightNodeEndpoint() string
}

type DAManager struct {
	datype config.DAType
	DataLayer
}

func NewDAManager(datype config.DAType, home string) *DAManager {
	return &DAManager{
		datype: datype,
		//FIXME: initiilize handler by type
		DataLayer: &celestia.Celestia{
			Root: home,
		},
	}
}
