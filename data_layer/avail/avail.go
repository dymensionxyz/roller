package avail

import (
	"fmt"
	"math/big"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"

	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	availtypes "github.com/centrifuge/go-substrate-rpc-client/v4/types"

	bip39 "github.com/cosmos/go-bip39"
)

const (
	availConfigFileName       = "avail.toml"
	mnemonicEntropySize       = 256
	keyringNetworkID    uint8 = 42
	DeafultRPCEndpoint        = "wss://dymension-devnet.avail.tools/ws"
	requiredAVL               = 1
)

type Avail struct {
	Root        string
	Mnemonic    string
	AccAddress  string
	RpcEndpoint string

	client *gsrpc.SubstrateAPI
}

func (a *Avail) GetPrivateKey() (string, error) {
	return a.Mnemonic, nil
}

func (a *Avail) SetMetricsEndpoint(endpoint string) {
}

func NewAvail(root string) *Avail {
	cfgPath := filepath.Join(root, availConfigFileName)
	availConfig, err := loadConfigFromTOML(cfgPath)
	if err != nil {
		entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
		if err != nil {
			panic(err)
		}

		availConfig.Mnemonic, err = bip39.NewMnemonic(entropySeed)
		if err != nil {
			panic(err)
		}

		err = writeConfigToTOML(cfgPath, availConfig)
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
		return nil, fmt.Errorf("failed to get DA balance: %w", err)
	}

	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	required := new(big.Int).Mul(big.NewInt(requiredAVL), exp)
	if required.Cmp(balance.Int) > 0 {
		return []utils.NotFundedAddressData{
			{
				KeyName:         a.GetKeyName(),
				Address:         a.AccAddress,
				CurrentBalance:  balance.Int,
				RequiredBalance: required,
				Denom:           consts.Denoms.Avail,
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
	if err != nil {
		return res, err
	}
	if !ok {
		return res, fmt.Errorf("account %s not found", keyringPair.Address)
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
				Denom:  consts.Denoms.Avail,
				Amount: balance.Int,
			},
		},
	}, nil
}

func (a *Avail) GetSequencerDAConfig() (string, error) {
	return fmt.Sprintf(`{"seed": "%s", "api_url": "%s", "app_id": 0, "tip":0}`, a.Mnemonic, a.RpcEndpoint), nil
}

func (a *Avail) SetRPCEndpoint(rpc string) {
	a.RpcEndpoint = rpc
}

func (a *Avail) GetLightNodeEndpoint() string {
	return ""
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
