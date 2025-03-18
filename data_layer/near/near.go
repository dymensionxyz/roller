package near

import (
	"os/exec"

	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

const (
	ConfigFileName = "near.toml"
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

type Near struct {
	Root        string
	PrivateKey  string
	RpcEndpoint string
	ChainID     uint32
}

func (w *Near) GetPrivateKey() (string, error) {
	return w.PrivateKey, nil
}

func (w *Near) SetMetricsEndpoint(endpoint string) {
}

func NewNear(root string) *Near {
	return nil
}

func (w *Near) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (w *Near) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (w *Near) GetRootDirectory() string {
	return w.Root
}

func (w *Near) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (w *Near) GetStartDACmd() *exec.Cmd {
	return nil
}

func (w *Near) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (w *Near) GetSequencerDAConfig(_ string) string {
	return ""
}

func (w *Near) SetRPCEndpoint(rpc string) {
	w.RpcEndpoint = rpc
}

func (w *Near) GetLightNodeEndpoint() string {
	return ""
}

func (w *Near) GetNetworkName() string {
	return "near"
}

func (w *Near) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (w *Near) GetKeyName() string {
	return "near"
}

func (w *Near) GetNamespaceID() string {
	return ""
}

func (w *Near) GetAppID() uint32 {
	return 0
}

func GetBalance(jsonRPCURL, key string) (string, error) {
	return "", nil
}
