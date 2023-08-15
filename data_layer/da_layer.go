package datalayer

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/data_layer/avail"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/data_layer/damock"
)

type DataLayer interface {
	GetDAAccountAddress() (string, error)
	InitializeLightNodeConfig() error
	CheckDABalance() ([]utils.NotFundedAddressData, error)
	GetStartDACmd() *exec.Cmd
	GetDAAccData(c config.RollappConfig) ([]utils.AccountData, error)
	GetLightNodeEndpoint() string
	GetSequencerDAConfig() string
	SetRPCEndpoint(string)
	SetMetricsEndpoint(endpoint string)
	GetNetworkName() string
	GetStatus(c config.RollappConfig) string
	GetKeyName() string
	GetPrivateKey() (string, error)
}

type DAManager struct {
	datype config.DAType
	DataLayer
}

func NewDAManager(datype config.DAType, home string) *DAManager {
	var dalayer DataLayer

	switch datype {
	case config.Celestia:
		dalayer = &celestia.Celestia{
			Root: home,
		}
	case config.Avail:
		dalayer = avail.NewAvail(home)
	case config.Mock:
		dalayer = &damock.DAMock{}
	default:
		panic("Unknown data layer type")
	}

	return &DAManager{
		datype:    datype,
		DataLayer: dalayer,
	}
}
