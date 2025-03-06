package avail

import (
	"fmt"
	"log"
	"math/big"
	"os/exec"

	cosmossdkmath "cosmossdk.io/math"
	gsrpc "github.com/centrifuge/go-substrate-rpc-client/v4"
	"github.com/centrifuge/go-substrate-rpc-client/v4/signature"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	prim "github.com/availproject/avail-go-sdk/primitives"
	"github.com/availproject/avail-go-sdk/sdk"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"

	syPallet "github.com/availproject/avail-go-sdk/metadata/pallets/system"
	"github.com/availproject/avail-go-sdk/primitives"
)

const (
	ConfigFileName             = "avail.toml"
	mnemonicEntropySize        = 256
	keyringNetworkID    uint16 = 42
	requiredAVL                = 1
)

type Avail struct {
	Root        string
	Mnemonic    string
	AccAddress  string
	RpcEndpoint string
	AppID       uint32

	client *gsrpc.SubstrateAPI
}

func (a *Avail) GetPrivateKey() (string, error) {
	return a.Mnemonic, nil
}

func (a *Avail) SetMetricsEndpoint(endpoint string) {
}

func NewAvail(root string) *Avail {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	availConfig, err := loadConfigFromTOML(cfgPath)

	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.AvailMainnet)
		} else {
			daNetwork = string(consts.AvailTestnet)
		}

		useExistingAvailWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to import an existing Avail wallet?",
		).Show()

		if useExistingAvailWallet {
			availConfig.Mnemonic, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your bip39 mnemonic",
			).Show()
		} else {
			entropySeed, err := bip39.NewEntropy(mnemonicEntropySize)
			if err != nil {
				panic(err)
			}

			availConfig.Mnemonic, err = bip39.NewMnemonic(entropySeed)
			if err != nil {
				panic(err)
			}

			fmt.Printf("\t%s\n", availConfig.Mnemonic)
			fmt.Println()
			fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))
		}

		keyringPair, err := signature.KeyringPairFromSecret(availConfig.Mnemonic, keyringNetworkID)
		if err != nil {
			panic(err)
		}

		pterm.DefaultSection.WithIndentCharacter("🔔").Println("Please fund your Avail addresses below")
		pterm.DefaultBasicText.Println(pterm.LightGreen(keyringPair.Address))

		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText(
				"press 'y' when the wallets are funded",
			).Show()

		if !proceed {
			panic(fmt.Errorf("Avail addr need to be fund!"))
		}

		insufficientBalances, err := availConfig.CheckDABalance()
		if err != nil {
			pterm.Error.Println("failed to check balance", err)
		}

		err = keys.PrintInsufficientBalancesIfAny(insufficientBalances)
		if err != nil {
			pterm.Error.Println("failed to check insufficient balances: ", err)
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			panic(fmt.Errorf("DA network configuration not found for: %s", daNetwork))
		}
		availConfig.RpcEndpoint = daData.ApiUrl
		availConfig.AccAddress = keyringPair.Address
		availConfig.Root = root

		availConfig.AppID, err = CreateAppID(rollerData.DA.ApiUrl, availConfig.Mnemonic, rollerData.RollappID)
		if err != nil {
			panic(err)
		}

		err = writeConfigToTOML(cfgPath, availConfig)
		if err != nil {
			panic(err)
		}

		pterm.Info.Printf("AppID: %d", availConfig.AppID)
	}
	return &availConfig
}

func (a *Avail) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (a *Avail) GetDAAccountAddress() (*keys.KeyInfo, error) {
	key := keys.KeyInfo{
		Address: a.AccAddress,
	}
	// return a.AccAddress, nil
	return &key, nil
}

func (a *Avail) GetRootDirectory() string {
	return a.Root
}

func (a *Avail) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	balance, err := a.getBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get DA balance: %w", err)
	}

	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(18), nil)
	required := new(big.Int).Mul(big.NewInt(requiredAVL), exp)
	if required.Cmp(balance) > 0 {
		return []keys.NotFundedAddressData{
			{
				KeyName:         a.GetKeyName(),
				Address:         a.AccAddress,
				CurrentBalance:  balance,
				RequiredBalance: required,
				Denom:           consts.Denoms.Avail,
				Network:         string(consts.Avail),
			},
		}, nil
	}
	return nil, nil
}

func (a *Avail) getBalance() (*big.Int, error) {
	// Initialize the SDK
	sdk, err := sdk.NewSDK(a.RpcEndpoint)
	if err != nil {
		log.Fatalf("Failed to initialize SDK: %v", err)
	}

	// Get the latest block hash
	latestBlockHash, err := sdk.Client.BestBlockHash()
	if err != nil {
		log.Fatalf("Failed to get latest block hash: %v", err)
	}

	// Initialize the block storage
	blockStorage, err := sdk.Client.StorageAt(prim.Some(latestBlockHash))
	if err != nil {
		log.Fatalf("Failed to get block storage: %v", err)
	}

	// Create the account ID
	accountId, err := primitives.NewAccountIdFromAddress(a.AccAddress)
	if err != nil {
		log.Fatalf("Failed to convert address: %v", err)
	}

	// Fetch the account data
	storage := syPallet.StorageAccount{}
	val, err := storage.Fetch(&blockStorage, accountId)
	if err != nil {
		log.Fatalf("Failed to fetch account: %v", err)
	}

	return val.Value.AccountData.Free.Value.Big(), nil
}

func (a *Avail) GetStartDACmd() *exec.Cmd {
	return nil
}

func (a *Avail) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	balance, err := a.getBalance()
	if err != nil {
		return nil, err
	}

	return []keys.AccountData{
		{
			Address: a.AccAddress,
			Balance: cosmossdktypes.Coin{
				Denom:  consts.Denoms.Avail,
				Amount: cosmossdkmath.NewIntFromBigInt(balance),
			},
		},
	}, nil
}

func (a *Avail) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"seed": "%s", "endpoint": "%s", "app_id": %d}`,
		a.Mnemonic,
		a.RpcEndpoint,
		a.AppID,
	)
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

func (a *Avail) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (a *Avail) GetKeyName() string {
	return "avail"
}

func (a *Avail) GetNamespaceID() string {
	return ""
}

func (a *Avail) GetAppID() uint32 {
	return a.AppID
}
