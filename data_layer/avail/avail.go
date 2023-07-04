package avail

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

const (
	rpcEndpoint = "wss://kate.avail.tools/ws"
)

type Avail struct {
	root     string
	mnemonic string
}

func NewAvail(root string) *Avail {
	return &Avail{
		root: root,
	}
}

func (a *Avail) GetDAAccountAddress() (string, error) {
	return "availtestaccount", nil
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

func (a *Avail) GetSequencerDAConfig() string {
	return fmt.Sprintf(`{"seed": "%s", "api_url": "%s", "app_id": 0, "tip":10}`, a.mnemonic, rpcEndpoint)
}
