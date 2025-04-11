package walrus

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"
)

const (
	ConfigFileName      = "walrus.toml"
	MnemonicEntropySize = 256
	requiredAVL         = 1
)

type Walrus struct {
	Root       string
	PrivateKey string
	Address    string
	Publisher  string
	Aggregator string
}

func (w *Walrus) GetPrivateKey() (string, error) {
	return w.PrivateKey, nil
}

func (w *Walrus) SetMetricsEndpoint(endpoint string) {
}

func NewWalrus(root string) *Walrus {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	walrusConfig, err := loadConfigFromTOML(cfgPath)

	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.WalrusMainnet)
		} else {
			daNetwork = string(consts.WalrusTestnet)
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			panic(fmt.Errorf("DA network configuration not found for: %b", daNetwork))
		}

		walrusConfig.Address, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"> Enter your blob owner address",
		).Show()

		walrusConfig.Publisher = daData.RpcUrl
		walrusConfig.Aggregator = daData.ApiUrl
		walrusConfig.Root = root

		err = writeConfigToTOML(cfgPath, walrusConfig)
		if err != nil {
			panic(err)
		}
	}
	return &walrusConfig
}

func (w *Walrus) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (w *Walrus) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (w *Walrus) GetRootDirectory() string {
	return w.Root
}

func (w *Walrus) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (w *Walrus) GetStartDACmd() *exec.Cmd {
	return nil
}

func (w *Walrus) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (w *Walrus) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"publisher_url": "%s", "aggregator_url": "%s", "blob_owner_addr": "%s", "store_duration_epochs": 180}`,
		w.Publisher,
		w.Aggregator,
		w.Address,
	)
}

func (w *Walrus) SetRPCEndpoint(rpc string) {
	w.Publisher = rpc
}

func (w *Walrus) GetLightNodeEndpoint() string {
	return ""
}

func (w *Walrus) GetNetworkName() string {
	return "walrus"
}

func (w *Walrus) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (w *Walrus) GetKeyName() string {
	return "walrus"
}

func (w *Walrus) GetNamespaceID() string {
	return ""
}

func (w *Walrus) GetAppID() uint32 {
	return 0
}
