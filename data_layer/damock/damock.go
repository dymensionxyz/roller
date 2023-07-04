package damock

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

// todo implemet data layer interface
type DAMock struct {
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

func (d *DAMock) GetStartDACmd(rpcEndpoint string) *exec.Cmd {
	return nil
}

func (d *DAMock) GetDAAccData(c config.RollappConfig) (*utils.AccountData, error) {
	return nil, nil
}

func (d *DAMock) GetLightNodeEndpoint() string {
	return ""
}
