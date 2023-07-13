package avail

import (
	"fmt"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
)

const (
	gatewayAddr             = "0.0.0.0"
	gatewayPort             = "26659"
	CelestiaRestApiEndpoint = "https://api-arabica-9.consensus.celestia-arabica.com"
	DefaultCelestiaRPC      = "consensus-full-arabica-9.celestia-arabica.com"
	DefaultCelestiaNetwork  = "arabica"
	DeafultRPCEndpoint      = "wss://kate.avail.tools/ws"
)

var ()

type Avail struct {
	root        string
	mnemonic    string
	rpcEndpoint string
}

func NewAvail(root string) *Avail {
	return &Avail{
		root:        root,
		mnemonic:    "",
		rpcEndpoint: DeafultRPCEndpoint,
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

func (a *Avail) GetStartDACmd() *exec.Cmd {
	//TODO: implement
	return nil
}

func (a *Avail) GetDAAccData(c config.RollappConfig) ([]utils.AccountData, error) {
	return nil, nil
}

func (a *Avail) GetLightNodeEndpoint() string {
	return ""
}

func (a *Avail) GetSequencerDAConfig() string {
	return fmt.Sprintf(`{"seed": "%s", "api_url": "%s", "app_id": 0, "tip":10}`, a.mnemonic, a.rpcEndpoint)
}

func (a *Avail) SetRPCEndpoint(rpc string) {
	a.rpcEndpoint = rpc
}

func (a *Avail) GetNetworkName() string {
	return "avail"
}

func (a *Avail) GetStatus(c config.RollappConfig) string {
	return ""
}

func (a *Avail) GetKeyName() string {
	return ""
}

func (a *Avail) GetExportKeyCmd() *exec.Cmd {
	return nil
}
