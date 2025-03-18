package sui

import (
	"os/exec"

	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

const (
	ConfigFileName = "sui.toml"
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

type Sui struct {
	Root        string
	PrivateKey  string
	RpcEndpoint string
	ChainID     uint32
}

func (w *Sui) GetPrivateKey() (string, error) {
	return w.PrivateKey, nil
}

func (w *Sui) SetMetricsEndpoint(endpoint string) {
}

func NewSui(root string) *Sui {
	return nil
}

func (w *Sui) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (w *Sui) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (w *Sui) GetRootDirectory() string {
	return w.Root
}

func (w *Sui) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (w *Sui) GetStartDACmd() *exec.Cmd {
	return nil
}

func (w *Sui) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (w *Sui) GetSequencerDAConfig(_ string) string {
	return ""
}

func (w *Sui) SetRPCEndpoint(rpc string) {
	w.RpcEndpoint = rpc
}

func (w *Sui) GetLightNodeEndpoint() string {
	return ""
}

func (w *Sui) GetNetworkName() string {
	return "sui"
}

func (w *Sui) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (w *Sui) GetKeyName() string {
	return "sui"
}

func (w *Sui) GetNamespaceID() string {
	return ""
}

func (w *Sui) GetAppID() uint32 {
	return 0
}

func GetBalance(jsonRPCURL, key string) (string, error) {
	return "", nil
}
