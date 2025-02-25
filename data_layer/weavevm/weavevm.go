package weavevm

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
	ConfigFileName               = "weavevm.toml"
	mnemonicEntropySize          = 256
	keyringNetworkID      uint16 = 42
	requiredAVL                  = 1
	DefaultTestnetChainID        = 9496
)

type WeaveVM struct {
	Root        string
	PrivateKey  string
	RpcEndpoint string
	ChainID     uint32
}

func (w *WeaveVM) GetPrivateKey() (string, error) {
	return w.PrivateKey, nil
}

func (w *WeaveVM) SetMetricsEndpoint(endpoint string) {
}

func NewWeaveVM(root string) *WeaveVM {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	weavevmConfig, err := loadConfigFromTOML(cfgPath)

	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.WeaveVMMainnet)
		} else {
			daNetwork = string(consts.WeaveVMTestnet)
		}

		weavevmConfig.PrivateKey, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"> Enter your PrivateKey without 0x",
		).Show()

		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText(
				"press 'y' when the wallet are funded",
			).Show()

		if !proceed {
			panic(fmt.Errorf("WeaveVM wallet need to be fund!"))
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			panic(fmt.Errorf("DA network configuration not found for: %s", daNetwork))
		}
		weavevmConfig.RpcEndpoint = daData.ApiUrl
		weavevmConfig.Root = root

		weavevmConfig.ChainID = DefaultTestnetChainID

		err = writeConfigToTOML(cfgPath, weavevmConfig)
		if err != nil {
			panic(err)
		}
	}
	return &weavevmConfig
}

func (w *WeaveVM) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (w *WeaveVM) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (w *WeaveVM) GetRootDirectory() string {
	return w.Root
}

func (w *WeaveVM) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (w *WeaveVM) GetStartDACmd() *exec.Cmd {
	return nil
}

func (w *WeaveVM) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (w *WeaveVM) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"endpoint": "%s", "chain_id": %d,"private_key_hex": %s}`,
		w.RpcEndpoint,
		w.ChainID,
		w.PrivateKey,
	)
}

func (w *WeaveVM) SetRPCEndpoint(rpc string) {
	w.RpcEndpoint = rpc
}

func (w *WeaveVM) GetLightNodeEndpoint() string {
	return ""
}

func (w *WeaveVM) GetNetworkName() string {
	return "weavevm"
}

func (w *WeaveVM) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (w *WeaveVM) GetKeyName() string {
	return "weavevm"
}

func (w *WeaveVM) GetNamespaceID() string {
	return ""
}

func (w *WeaveVM) GetAppID() uint32 {
	return 0
}
