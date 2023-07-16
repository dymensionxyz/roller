package avail

import (
	"fmt"
	"io/ioutil"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/pelletier/go-toml"

	bip39 "github.com/cosmos/go-bip39"
)

const (
	availConfigFileName     = "avail.toml"
	mnemonicEntropySize     = 256
	gatewayAddr             = "0.0.0.0"
	gatewayPort             = "26659"
	CelestiaRestApiEndpoint = "https://api-arabica-9.consensus.celestia-arabica.com"
	DefaultCelestiaRPC      = "consensus-full-arabica-9.celestia-arabica.com"
	DefaultCelestiaNetwork  = "arabica"
	DeafultRPCEndpoint      = "wss://kate.avail.tools/ws"
)

type Avail struct {
	Root        string
	Mnemonic    string
	RpcEndpoint string
}

func NewAvail(root string) *Avail {
	cfgPath := filepath.Join(root, availConfigFileName)
	availConfig, err := LoadConfigFromTOML(cfgPath)
	if err != nil {
		fmt.Println("avail config not found, creating new mnemonic")
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			panic(err)
		}

		availConfig.Mnemonic, err = bip39.NewMnemonic(entropySeed)
		if err != nil {
			panic(err)
		}

		err = WriteConfigToTOML(cfgPath, availConfig)
		if err != nil {
			panic(err)
		}
	}

	availConfig.Root = root
	availConfig.RpcEndpoint = DeafultRPCEndpoint
	return &availConfig
}

func (a *Avail) InitializeLightNodeConfig() error {
	return nil
}

func (a *Avail) GetDAAccountAddress() (string, error) {
	//TODO: get address instead of mnemonic.
	//we should be able to get the address from the mnemonic using avail's KeyringPairFromSecret()
	return a.Mnemonic, nil
}

func (a *Avail) CheckDABalance() ([]utils.NotFundedAddressData, error) {
	//TODO: implement
	return nil, nil
}

func (a *Avail) GetStartDACmd() *exec.Cmd {
	return nil
}

func (a *Avail) GetDAAccData(c config.RollappConfig) ([]utils.AccountData, error) {
	//TODO: implement
	return nil, nil
}

func (a *Avail) GetLightNodeEndpoint() string {
	return a.RpcEndpoint
}

func (a *Avail) GetSequencerDAConfig() string {
	return fmt.Sprintf(`{"seed": "%s", "api_url": "%s", "app_id": 0, "tip":0}`, a.Mnemonic, a.RpcEndpoint)
}

func (a *Avail) SetRPCEndpoint(rpc string) {
	a.RpcEndpoint = rpc
}

func (a *Avail) GetNetworkName() string {
	return "avail"
}

func (a *Avail) GetStatus(c config.RollappConfig) string {
	return ""
}

func (a *Avail) GetKeyName() string {
	return "avail"
}

// FIXME: currently can't export the key from avail
func (a *Avail) GetExportKeyCmd() *exec.Cmd {
	return nil
}

/* -------------------------------------------------------------------------- */
/*                                    utils                                   */
/* -------------------------------------------------------------------------- */

// FIXME: config package should be refactored so this could be reused
func WriteConfigToTOML(path string, c Avail) error {
	tomlBytes, err := toml.Marshal(c)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, tomlBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}

func LoadConfigFromTOML(path string) (Avail, error) {
	var config Avail
	tomlBytes, err := ioutil.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = toml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
}
