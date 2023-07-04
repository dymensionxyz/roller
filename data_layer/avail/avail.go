package avail

import (
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

type Avail struct {
	root string
}

func NewAvail(root string) *Avail {
	return &Avail{
		root: root,
	}
}

func (a *Avail) GetDAAccountAddress() (string, error) {
	return "", nil
}

func (a *Avail) InitializeLightNodeConfig() error {
	return nil
}

func (a *Avail) CheckDABalance() ([]utils.NotFundedAddressData, error) {
	return nil, nil
}

func (a *Avail) GetStartDACmd(rpcEndpoint string) *exec.Cmd {
	return nil
}

func (a *Avail) GetDAAccData(c config.RollappConfig) ([]utils.AccountData, error) {
	return nil, nil
}

func (a *Avail) GetLightNodeEndpoint() string {
	return ""
}
