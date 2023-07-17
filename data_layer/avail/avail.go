package avail

import (
	"fmt"
	"io/ioutil"
	"math/big"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/pelletier/go-toml"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	availtypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"

	bip39 "github.com/cosmos/go-bip39"
)

const (
	availConfigFileName       = "avail.toml"
	mnemonicEntropySize       = 256
	keyringNetworkID    uint8 = 42
	DeafultRPCEndpoint        = "wss://kate.avail.tools/ws"
	requiredAVL               = 1
)

type Avail struct {
	Root        string
	Mnemonic    string
	AccAddress  string
	RpcEndpoint string

	client *gsrpc.SubstrateAPI
	accKey *availtypes.StorageKey
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

	keyringPair, err := signature.KeyringPairFromSecret(availConfig.Mnemonic, keyringNetworkID)
	if err != nil {
		panic(err)
	}
	availConfig.AccAddress = keyringPair.Address

	availConfig.Root = root
	availConfig.RpcEndpoint = DeafultRPCEndpoint
	return &availConfig
}

func (a *Avail) InitializeLightNodeConfig() error {
	return nil
}

func (a *Avail) GetDAAccountAddress() (string, error) {
	return a.AccAddress, nil
}
func (a *Avail) CheckDABalance() ([]utils.NotFundedAddressData, error) {
	balance, err := a.getBalance()
	if err != nil {
		return nil, err
	}

	fmt.Println("Balance: ", balance.Int.String())

	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	required := new(big.Int).Mul(big.NewInt(requiredAVL), exp)

	if balance.Int.Cmp(required) < 0 {
		return []utils.NotFundedAddressData{
			{
				KeyName:         a.GetKeyName(),
				Address:         a.AccAddress,
				CurrentBalance:  balance.Int,
				RequiredBalance: required,
				Denom:           "aAVL",
				Network:         "avail",
			},
		}, nil
	}
	return nil, nil
}

func (a *Avail) getBalance() (availtypes.U128, error) {
	if a.client == nil {
		client, err := gsrpc.NewSubstrateAPI(DeafultRPCEndpoint)
		if err != nil {
			return availtypes.U128{}, err
		}
		a.client = client
	}
	var res availtypes.U128
	meta, err := a.client.RPC.State.GetMetadataLatest()
	if err != nil {
		return res, err
	}

	keyringPair, err := signature.KeyringPairFromSecret(a.Mnemonic, keyringNetworkID)
	if err != nil {
		return res, err
	}
	key, err := availtypes.CreateStorageKey(meta, "System", "Account", keyringPair.PublicKey)
	if err != nil {
		return res, err
	}

	var accountInfo availtypes.AccountInfo
	ok, err := a.client.RPC.State.GetStorageLatest(key, &accountInfo)
	if err != nil || !ok {
		return res, err
	}

	return accountInfo.Data.Free, nil
}

func (a *Avail) GetStartDACmd() *exec.Cmd {
	return nil
}

func (a *Avail) GetDAAccData(c config.RollappConfig) ([]utils.AccountData, error) {
	balance, err := a.getBalance()
	if err != nil {
		return nil, err
	}
	return []utils.AccountData{
		{
			Address: a.AccAddress,
			Balance: utils.Balance{
				//FIXME: denom should be AVL. use huminize or utils.balance
				Denom:  "aAVL",
				Amount: balance.Int,
			},
		},
	}, nil
}

func (a *Avail) GetSequencerDAConfig() string {
	return fmt.Sprintf(`{"seed": "%s", "api_url": "%s", "app_id": 0, "tip":0}`, a.Mnemonic, a.RpcEndpoint)
}

func (a *Avail) SetRPCEndpoint(rpc string) {
	a.RpcEndpoint = rpc
}

func (a *Avail) GetLightNodeEndpoint() string {
	return a.RpcEndpoint
}

func (a *Avail) GetNetworkName() string {
	return "avail"
}

func (a *Avail) GetStatus(c config.RollappConfig) string {
	return "Running"
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
