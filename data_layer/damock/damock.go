package damock

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

type DAMock struct {
}

func (d *DAMock) GetStatus(c config.RollappConfig) string {
	return "Running mock DA"
}

func (d *DAMock) GetExportKeyCmd() *exec.Cmd {
	return nil
}

func NewDAMock() *DAMock {
	return &DAMock{}
}

func (d *DAMock) GetDAAccountAddress() (string, error) {
	return "", nil
}

func (d *DAMock) InitializeLightNodeConfig() error {
	return nil
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
	return "mock"
}
