package datalayer

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/data_layer/aptos"
	"github.com/dymensionxyz/roller/data_layer/avail"
	"github.com/dymensionxyz/roller/data_layer/bnb"
	"github.com/dymensionxyz/roller/data_layer/celestia"
	"github.com/dymensionxyz/roller/data_layer/damock"
	loadnetwork "github.com/dymensionxyz/roller/data_layer/loadnetwork"
	"github.com/dymensionxyz/roller/data_layer/sui"
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
	DaType consts.DAType
	DataLayer
}

func NewDAManager(datype consts.DAType, home string, kb consts.SupportedKeyringBackend) *DAManager {
	var dalayer DataLayer

	switch datype {
	case consts.Celestia:
		dalayer = celestia.NewCelestia(home, kb)
	case consts.Avail:
		dalayer = avail.NewAvail(home)
	case consts.Aptos:
		dalayer = aptos.NewAptos(home)
	case consts.Sui:
		dalayer = sui.NewSui(home)
	case consts.LoadNetwork:
		dalayer = loadnetwork.NewLoadNetwork(home)
	case consts.Bnb:
		dalayer = bnb.NewBnb(home)
	case consts.Local:
		dalayer = &damock.DAMock{}
	default:
		panic("Unknown data layer type " + string(datype))
	}

	return &DAManager{
		DaType:    datype,
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
		case string(consts.Aptos):
			daNetwork = string(consts.AptosTestnet)
		case string(consts.Sui):
			daNetwork = string(consts.SuiTestnet)
		case string(consts.LoadNetwork):
			daNetwork = string(consts.LoadNetworkTestnet)
		case string(consts.Bnb):
			daNetwork = string(consts.BnbTestnet)
		default:
			return nil, fmt.Errorf("unsupported DA backend: %s", daBackend)
		}
	case "mainnet":
		switch daBackend {
		case string(consts.Celestia):
			daNetwork = string(consts.CelestiaMainnet)
		case string(consts.Avail):
			daNetwork = string(consts.AvailMainnet)
		case string(consts.Aptos):
			daNetwork = string(consts.AptosMainnet)
		case string(consts.Sui):
			daNetwork = string(consts.SuiMainnet)
		case string(consts.LoadNetwork):
			daNetwork = string(consts.LoadNetworkMainnet)
		case string(consts.Bnb):
			daNetwork = string(consts.BnbMainnet)
		default:
			return nil, fmt.Errorf("unsupported DA backend: %s", daBackend)
		}
	case "custom":
		switch daBackend {
		case string(consts.Celestia):
			daNetwork = string(consts.CelestiaTestnet)
		case string(consts.Avail):
			daNetwork = string(consts.AvailTestnet)
		case string(consts.Aptos):
			daNetwork = string(consts.AptosTestnet)
		case string(consts.Sui):
			daNetwork = string(consts.SuiTestnet)
		case string(consts.LoadNetwork):
			daNetwork = string(consts.LoadNetworkTestnet)
		case string(consts.Bnb):
			daNetwork = string(consts.BnbTestnet)
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
