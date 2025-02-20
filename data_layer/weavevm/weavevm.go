package weavevm

import (
	"os/exec"

	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

type WeaveVM struct {
	PrivateKey  string
	RpcEndpoint string
	ChainId     uint32
}

func (wv *WeaveVM) GetPrivateKey() (string, error) {
	return wv.PrivateKey, nil
}

func (wv *WeaveVM) SetMetricsEndpoint(endpoint string) {
}

func (wv *WeaveVM) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (wv *WeaveVM) GetRootDirectory() string {
	return ""
}

func (wv *WeaveVM) GetNamespaceID() string {
	return ""
}

func NewWeaveVM() *WeaveVM {
	return &WeaveVM{}
}

func (wv *WeaveVM) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (wv *WeaveVM) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (wv *WeaveVM) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return []keys.NotFundedAddressData{}, nil
}

func (wv *WeaveVM) GetStartDACmd() *exec.Cmd {
	return nil
}

func (wv *WeaveVM) GetDAAccData(c roller.RollappConfig) ([]keys.AccountData, error) {
	return []keys.AccountData{}, nil
}

func (wv *WeaveVM) GetLightNodeEndpoint() string {
	return ""
}

func (wv *WeaveVM) GetSequencerDAConfig(nt string) string {
	return ""
}

func (wv *WeaveVM) SetRPCEndpoint(string) {
}

func (wv *WeaveVM) GetKeyName() string {
	return ""
}

func (wv *WeaveVM) GetNetworkName() string {
	return "local"
}

func (wv *WeaveVM) GetAppID() uint32 {
	return 0
}
