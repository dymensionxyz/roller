package datalayer

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/data_layer/avail"
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
	GetAppID() uint32
}

type DAManager struct {
	datype consts.DAType
	DataLayer
}

func NewDAManager(datype consts.DAType, home string, kb consts.SupportedKeyringBackend) *DAManager {
	var dalayer DataLayer

	switch datype {
	case consts.Celestia:
		dalayer = celestia.NewCelestia(home, kb)
	case consts.Avail:
		dalayer = avail.NewAvail(home)
	case consts.Local:
		dalayer = &damock.DAMock{}
	default:
		panic("Unknown data layer type " + string(datype))
	}

	return &DAManager{
		datype:    datype,
		DataLayer: dalayer,
	}
}

func GetDaInfo(env, daBackend string) (*consts.DaData, error) {
	var daNetwork string

	switch env {
	case "playground", "blumbus":
		switch daBackend {
		case string(consts.Celestia):
			daNetwork = string(consts.CelestiaTestnet)
		case string(consts.Avail):
			daNetwork = string(consts.AvailTestnet)
		default:
			return nil, fmt.Errorf("unsupported DA backend: %s", daBackend)
		}
	case "mainnet":
		switch daBackend {
		case string(consts.Celestia):
			daNetwork = string(consts.CelestiaMainnet)
		case string(consts.Avail):
			daNetwork = string(consts.AvailMainnet)
		default:
			return nil, fmt.Errorf("unsupported DA backend: %s", daBackend)
		}
	case "custom":
		switch daBackend {
		case string(consts.Celestia):
			daNetwork = string(consts.CelestiaTestnet)
		case string(consts.Avail):
			daNetwork = string(consts.AvailTestnet)
		default:
			return nil, fmt.Errorf("unsupported DA backend: %s", daBackend)
		}
	case "mock":
		daNetwork = "mock"
	default:
		return nil, fmt.Errorf("unsupported environment: %s", env)
	}

	// Check if the daNetwork exists in DaNetworks
	daData, exists := consts.DaNetworks[daNetwork]
	if !exists {
		return nil, fmt.Errorf("DA network configuration not found for: %s", daNetwork)
	}

	return &daData, nil
}
