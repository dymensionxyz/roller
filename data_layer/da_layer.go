package datalayer

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/data_layer/damock"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

type DataLayer interface {
	GetDAAccountAddress() (*keys.KeyInfo, error)
	InitializeLightNodeConfig() (string, error)
	CheckDABalance() ([]keys.NotFundedAddressData, error)
	GetStartDACmd() *exec.Cmd
	GetDAAccData(c roller.RollappConfig) ([]keys.AccountData, error)
	GetLightNodeEndpoint() string
	// todo: Refactor, node type makes reusability awful
	GetSequencerDAConfig(nt string) string
	SetRPCEndpoint(string)
	SetMetricsEndpoint(endpoint string)
	GetNetworkName() string
	GetStatus(c roller.RollappConfig) string
	GetKeyName() string
	GetPrivateKey() (string, error)
	GetRootDirectory() string
	GetNamespaceID() string
}

type DAManager struct {
	datype consts.DAType
	DataLayer
}

func NewDAManager(datype consts.DAType, home string) *DAManager {
	var dalayer DataLayer

	switch datype {
	case consts.Celestia:
		dalayer = celestia.NewCelestia(home)
	// case config.Avail:
	// 	dalayer = avail.NewAvail(home)
	case consts.Local:
		dalayer = &damock.DAMock{}
	default:
		panic("Unknown data layer type")
	}

	return &DAManager{
		datype:    datype,
		DataLayer: dalayer,
	}
}
