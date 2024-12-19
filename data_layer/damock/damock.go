package damock

import (
	"os/exec"

	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
)

type DAMock struct{}

func (d *DAMock) GetPrivateKey() (string, error) {
	return "", nil
}

func (d *DAMock) SetMetricsEndpoint(endpoint string) {
}

func (d *DAMock) GetStatus(c roller.RollappConfig) string {
	return "Running local DA"
}

func (d *DAMock) GetRootDirectory() string {
	return ""
}

func (d *DAMock) GetNamespaceID() string {
	return ""
}

func NewDAMock() *DAMock {
	return &DAMock{}
}

func (d *DAMock) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (d *DAMock) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (d *DAMock) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return []keys.NotFundedAddressData{}, nil
}

func (d *DAMock) GetStartDACmd() *exec.Cmd {
	return nil
}

func (d *DAMock) GetDAAccData(c roller.RollappConfig) ([]keys.AccountData, error) {
	return []keys.AccountData{}, nil
}

func (d *DAMock) GetLightNodeEndpoint() string {
	return ""
}

func (d *DAMock) GetSequencerDAConfig(nt string) string {
	return ""
}

func (d *DAMock) SetRPCEndpoint(string) {
}

func (d *DAMock) GetKeyName() string {
	return ""
}

func (d *DAMock) GetNetworkName() string {
	return "local"
}

func (d *DAMock) GetAppID() int {
	return 0
}
