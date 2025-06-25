package ethereum

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
	ConfigFileName = "ethereum.toml"
)

type Ethereum struct {
	Root        string
	PrivateKey  string
	Address     string
	RpcEndpoint string
	ApiEndpoint string
	GasLimit    int
	ChainID     int
}

func (e *Ethereum) GetPrivateKey() (string, error) {
	return e.PrivateKey, nil
}

func (e *Ethereum) SetMetricsEndpoint(endpoint string) {}

func NewEthereum(root string) *Ethereum {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	ethConfig, err := loadConfigFromTOML(cfgPath)
	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.EthereumMainnet)
			ethConfig.ChainID = 1
		} else {
			daNetwork = string(consts.EthereumTestnet)
			ethConfig.ChainID = 11155111
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			pterm.Error.Printf("DA network configuration not found for: %s", daNetwork)
			return &ethConfig
		}

		ethConfig.RpcEndpoint = daData.RpcUrl
		ethConfig.Root = root
		ethConfig.GasLimit = 100000

		useExistingRpcEndpoint, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to use your own RPC endpoint??",
		).Show()

		if useExistingRpcEndpoint {
			ethConfig.RpcEndpoint, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your RPC endpoint",
			).Show()
		} else {
			ethConfig.RpcEndpoint = daData.RpcUrl
		}

		useExistingAPIEndpoint, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to use your own API endpoint??",
		).Show()

		if useExistingAPIEndpoint {
			ethConfig.ApiEndpoint, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your API endpoint",
			).Show()
		} else {
			ethConfig.ApiEndpoint = daData.ApiUrl
		}

		pterm.Warning.Print("You will need to save Eth Private Key to an environment variable named ETH_PRIVATE_KEY")

		err = writeConfigToTOML(cfgPath, ethConfig)
		if err != nil {
			panic(err)
		}
	}
	return &ethConfig
}

func (e *Ethereum) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (e *Ethereum) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (e *Ethereum) GetRootDirectory() string {
	return e.Root
}

func (e *Ethereum) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (e *Ethereum) GetStartDACmd() *exec.Cmd {
	return nil
}

func (e *Ethereum) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (e *Ethereum) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"endpoint":"%s","gas_limit":%d,"private_key_env":"ETH_PRIVATE_KEY","chain_id":%d,"api_url":"%s"}`,
		e.RpcEndpoint,
		e.GasLimit,
		e.ChainID,
		e.ApiEndpoint,
	)
}

func (e *Ethereum) SetRPCEndpoint(rpc string) {
	e.RpcEndpoint = rpc
}

func (e *Ethereum) GetLightNodeEndpoint() string {
	return ""
}

func (e *Ethereum) GetNetworkName() string {
	return "ethereum"
}

func (e *Ethereum) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (e *Ethereum) GetKeyName() string {
	return "ethereum"
}

func (e *Ethereum) GetNamespaceID() string {
	return ""
}

func (e *Ethereum) GetAppID() uint32 {
	return 0
}
