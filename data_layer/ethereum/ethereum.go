package ethereum

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

const (
	ConfigFileName = "ethereum.toml"
)

type Ethereum struct {
	Root          string
	PrivateKey    string
	Address       string
	Publisher     string
	Aggregator    string
	Endpoint      string
	GasLimit      int
	PrivateKeyEnv string
	ChainID       int
	ApiUrl        string
}

func (e *Ethereum) GetPrivateKey() (string, error) {
	return e.PrivateKey, nil
}

func (e *Ethereum) SetMetricsEndpoint(endpoint string) {}

func NewEthereum(root string) *Ethereum {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	if err != nil {
		return &Ethereum{Root: root}
	}

	if rollerData.HubData.Environment == "mainnet" {
		daNetwork = string(consts.EthereumMainnet)
	} else {
		daNetwork = string(consts.EthereumTestnet)
	}

	daData, exists := consts.DaNetworks[daNetwork]
	if !exists {
		return &Ethereum{Root: root}
	}

	return &Ethereum{
		Root:       root,
		Publisher:  daData.RpcUrl,
		Aggregator: daData.ApiUrl,
	}
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
		`{"endpoint":"%s","gas_limit":%d,"private_key_env":"%s","chain_id":%d,"api_url":"%s"}`,
		e.Endpoint,
		e.GasLimit,
		e.PrivateKeyEnv,
		e.ChainID,
		e.ApiUrl,
	)
}

func (e *Ethereum) SetRPCEndpoint(rpc string) {
	e.Publisher = rpc
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
