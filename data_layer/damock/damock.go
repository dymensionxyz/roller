package damock

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

type DAMock struct{}

func (d *DAMock) GetPrivateKey() (string, error) {
	return "", nil
}

func (d *DAMock) SetMetricsEndpoint(endpoint string) {
}

func (d *DAMock) GetStatus(c config.RollappConfig) string {
	return "Running local DA"
}

func (c *DAMock) GetRootDirectory() string {
	return ""
}

func (c *DAMock) GetNamespaceID() string {
	return ""
}

func NewDAMock() *DAMock {
	return &DAMock{}
}

func (d *DAMock) GetDAAccountAddress() (*utils.KeyInfo, error) {
	return nil, nil
}

func (d *DAMock) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (d *DAMock) CheckDABalance() ([]utils.NotFundedAddressData, error) {
	return []utils.NotFundedAddressData{}, nil
}

func (d *DAMock) GetStartDACmd() *exec.Cmd {
	return nil
}

func (d *DAMock) GetDAAccData(c config.RollappConfig) ([]utils.AccountData, error) {
	return []utils.AccountData{}, nil
}

func (d *DAMock) GetLightNodeEndpoint() string {
	return ""
}

func (d *DAMock) GetSequencerDAConfig() string {
	return ""
}

func (d *DAMock) SetRPCEndpoint(string) {
}

func (c *DAMock) GetKeyName() string {
	return ""
}

func (d *DAMock) GetNetworkName() string {
	return "local"
}
