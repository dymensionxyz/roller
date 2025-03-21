package aptos

import (
	"os/exec"

	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

const (
	ConfigFileName = "aptos.toml"
)

type RequestPayload struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type EthBalanceResponse struct {
	ID      int    `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  string `json:"result"`
}

type Aptos struct {
	Root        string
	PrivateKey  string
	RpcEndpoint string
	ChainID     uint32
}

func (a *Aptos) GetPrivateKey() (string, error) {
	return a.PrivateKey, nil
}

func (a *Aptos) SetMetricsEndpoint(endpoint string) {
}

func NewAptos(root string) *Aptos {
	return nil
}

func (a *Aptos) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (a *Aptos) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (a *Aptos) GetRootDirectory() string {
	return a.Root
}

func (a *Aptos) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (a *Aptos) GetStartDACmd() *exec.Cmd {
	return nil
}

func (a *Aptos) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (a *Aptos) GetSequencerDAConfig(_ string) string {
	return ""
}

func (a *Aptos) SetRPCEndpoint(rpc string) {
	a.RpcEndpoint = rpc
}

func (a *Aptos) GetLightNodeEndpoint() string {
	return ""
}

func (a *Aptos) GetNetworkName() string {
	return "aptos"
}

func (a *Aptos) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (a *Aptos) GetKeyName() string {
	return "aptos"
}

func (a *Aptos) GetNamespaceID() string {
	return ""
}

func (a *Aptos) GetAppID() uint32 {
	return 0
}

func GetBalance(jsonRPCURL, key string) (string, error) {
	return "", nil
}
