package aptos

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"os/exec"

	"cosmossdk.io/math"
	aptos "github.com/aptos-labs/aptos-go-sdk"
	"github.com/aptos-labs/aptos-go-sdk/crypto"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"
)

const (
	ConfigFileName = "aptos.toml"
)

type RequestPayload struct {
	JSONRPC string        `json:"jsonrpc"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
	ID      int           `json:"id"`
}

type EthBalanceResponse struct {
	ID      int    `json:"id"`
	JsonRPC string `json:"jsonrpc"`
	Result  string `json:"result"`
}

type Aptos struct {
	Root        string
	PrivateKey  string
	RpcEndpoint string
	Network     string
}

func (a *Aptos) GetPrivateKey() (string, error) {
	return a.PrivateKey, nil
}

func (a *Aptos) SetMetricsEndpoint(endpoint string) {
}

type Ed25519PrivateKey struct {
	Inner ed25519.PrivateKey // Inner is the actual private key
}

func NewAptos(root string) *Aptos {
	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	aptConfig, err := LoadConfigFromTOML(cfgPath)

	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			aptConfig.Network = string(consts.AptosMainnet)
		} else {
			aptConfig.Network = string(consts.AptosTestnet)
		}

		daData, exists := consts.DaNetworks[aptConfig.Network]
		if !exists {
			panic(fmt.Errorf("DA network configuration not found for: %s", aptConfig.Network))
		}

		useExistingAPTWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to import an existing APT wallet?",
		).Show()

		if useExistingAPTWallet {
			aptConfig.PrivateKey, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your APT private key",
			).Show()
		} else {
			_, privKey, err := ed25519.GenerateKey(nil)
			if err != nil {
				panic(err)
			}

			privKeyHex := hex.EncodeToString(privKey.Seed())
			aptConfig.PrivateKey = privKeyHex
			if err != nil {
				panic(err)
			}

			fmt.Printf("\t%s\n", aptConfig.PrivateKey)
			fmt.Println()
			fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))
		}

		pterm.DefaultSection.WithIndentCharacter("🔔").Println("Please fund your APT wallet have privatekey below")
		pterm.DefaultBasicText.Println(pterm.LightGreen(aptConfig.PrivateKey))

		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText(
				"press 'y' when the wallets are funded",
			).Show()

		if !proceed {
			panic(fmt.Errorf("APT addr need to be fund!"))
		}

		aptConfig.RpcEndpoint = daData.RpcUrl
		aptConfig.Root = root

		balance, err := aptConfig.getBalance()
		if err != nil {
			panic(err)
		}

		if balance == 0 {
			panic(fmt.Errorf("APT wallet need to be fund!"))
		}

		pterm.Println("APT Balance: ", balance)

		pterm.Warning.Print("You will need to save Private Key to an environment variable named APT_PRIVATE_KEY")

		err = writeConfigToTOML(cfgPath, aptConfig)
		if err != nil {
			panic(err)
		}
	}
	return &aptConfig
}

func (a *Aptos) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (a *Aptos) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (a *Aptos) GetRootDirectory() string {
	return a.Root
}

func (a *Aptos) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (a *Aptos) GetStartDACmd() *exec.Cmd {
	return nil
}

func (a *Aptos) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	balance, err := a.getBalance()
	if err != nil {
		return nil, err
	}

	return []keys.AccountData{
		{
			Address: a.PrivateKey,
			Balance: cosmossdktypes.Coin{
				Denom:  consts.Denoms.Aptos,
				Amount: math.NewIntFromUint64(balance),
			},
		},
	}, nil
}

func (a *Aptos) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"network": "%s", "pri_key_env": "APT_PRIVATE_KEY"}`,
		a.Network,
	)
}

func (a *Aptos) SetRPCEndpoint(rpc string) {
	a.RpcEndpoint = rpc
}

func (a *Aptos) GetLightNodeEndpoint() string {
	return ""
}

func (a *Aptos) GetNetworkName() string {
	return "aptos"
}

func (a *Aptos) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (a *Aptos) GetKeyName() string {
	return "aptos"
}

func (a *Aptos) GetNamespaceID() string {
	return ""
}

func (a *Aptos) GetAppID() uint32 {
	return 0
}

func (a *Aptos) getBalance() (uint64, error) {
	var client *aptos.Client
	var err error

	if a.Network == "mainnet" {
		client, err = aptos.NewClient(aptos.MainnetConfig)
		if err != nil {
			return 0, err
		}
	} else {
		client, err = aptos.NewClient(aptos.TestnetConfig)
		if err != nil {
			return 0, err
		}
	}
	key := crypto.Ed25519PrivateKey{}
	err = key.FromHex(a.PrivateKey)
	if err != nil {
		return 0, err
	}
	acc, err := aptos.NewAccountFromSigner(&key)
	if err != nil {
		return 0, err
	}
	balance, err := client.AccountAPTBalance(acc.Address)
	if err != nil {
		return 0, err
	}
	return balance, nil
}
