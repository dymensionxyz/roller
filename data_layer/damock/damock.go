package damock

import (
	"math/big"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

// todo implemet data layer interface
type DAMock struct {
}

func (d *DAMock) GetStatus(c config.RollappConfig) string {
	//TODO implement me
	return ""
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
	return []utils.AccountData{
		{
			Address: "mockDA",
			Balance: utils.Balance{
				Denom:  "mock",
				Amount: big.NewInt(999999999999999),
			},
		},
	}, nil
}

func (d *DAMock) GetLightNodeEndpoint() string {
	return ""
}

func (d *DAMock) SetRPCEndpoint(string) {
}

func (d *DAMock) GetNetworkName() string {
	return "mock"
}
