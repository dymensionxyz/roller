package sui

import (
	"context"
	"fmt"
	"math/big"
	"os/exec"

	"github.com/block-vision/sui-go-sdk/models"
	"github.com/block-vision/sui-go-sdk/signer"
	"github.com/block-vision/sui-go-sdk/sui"
	"github.com/cosmos/go-bip39"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"
)

const (
	ConfigFileName             = "sui.toml"
	DefaultTestnetChainID      = 9496
	NoopContractAddressTestnet = "0xcf119583badb169bfc9a031ec16fb6a79a5151ff7aa0d229f2a35b798ddcd9d6"
	NoopContractAddressMainnet = "0x015596db61510363c3341b8a4ede0986fadcb1b5bed70d2586b8e2e7db9538a7"
	MnemonicEntropySize        = 256
	requiredAVL                = 1
)

type Sui struct {
	Root                string
	Mnemonic            string
	Address             string
	NoopContractAddress string
	RpcEndpoint         string
	ChainID             uint32
}

func (s *Sui) GetPrivateKey() (string, error) {
	return s.Mnemonic, nil
}

func (s *Sui) SetMetricsEndpoint(endpoint string) {
}

func NewSui(root string) *Sui {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	suiConfig, err := LoadConfigFromTOML(cfgPath)

	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.SuiMainnet)
			suiConfig.NoopContractAddress = NoopContractAddressMainnet
		} else {
			daNetwork = string(consts.SuiTestnet)
			suiConfig.NoopContractAddress = NoopContractAddressTestnet
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			panic(fmt.Errorf("DA network configuration not found for: %s", daNetwork))
		}

		useExistingSuiWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to import an existing SUI wallet?",
		).Show()

		if useExistingSuiWallet {
			suiConfig.Mnemonic, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your bip39 mnemonic",
			).Show()
		} else {
			entropySeed, err := bip39.NewEntropy(MnemonicEntropySize)
			if err != nil {
				panic(err)
			}

			suiConfig.Mnemonic, err = bip39.NewMnemonic(entropySeed)
			if err != nil {
				panic(err)
			}

			fmt.Printf("\t%s\n", suiConfig.Mnemonic)
			fmt.Println()
			fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))
		}

		key, err := signer.NewSignertWithMnemonic(suiConfig.Mnemonic)
		if err != nil {
			panic(err)
		}

		pterm.DefaultSection.WithIndentCharacter("🔔").Println("Please fund your sui addresses below")
		pterm.DefaultBasicText.Println(pterm.LightGreen(key.Address))

		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText(
				"press 'y' when the wallets are funded",
			).Show()

		if !proceed {
			panic(fmt.Errorf("Sui addr need to be fund!"))
		}

		suiConfig.RpcEndpoint = daData.RpcUrl
		suiConfig.Root = root
		suiConfig.Address = key.Address

		insufficientBalances, err := suiConfig.CheckDABalance()
		if err != nil {
			pterm.Error.Println("failed to check balance", err)
		}

		err = keys.PrintInsufficientBalancesIfAny(insufficientBalances)
		if err != nil {
			pterm.Error.Println("failed to check insufficient balances: ", err)
		}

		err = writeConfigToTOML(cfgPath, suiConfig)
		if err != nil {
			panic(err)
		}

		pterm.Warning.Print("You will need to save Mnemonic to an environment variable named SUI_MNEMONIC")
	}
	return &suiConfig
}

func (s *Sui) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (s *Sui) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (s *Sui) GetRootDirectory() string {
	return s.Root
}

func (s *Sui) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	balance, err := s.getBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get DA balance: %w", err)
	}

	exp := new(big.Int).Exp(big.NewInt(1000), big.NewInt(1), nil)
	required := new(big.Int).Mul(big.NewInt(requiredAVL), exp)
	if required.Cmp(balance) > 0 {
		return []keys.NotFundedAddressData{
			{
				KeyName:         s.GetKeyName(),
				Address:         s.Address,
				CurrentBalance:  balance,
				RequiredBalance: required,
				Denom:           consts.Denoms.Sui,
				Network:         string(consts.Sui),
			},
		}, nil
	}
	return nil, nil
}

func (s *Sui) GetStartDACmd() *exec.Cmd {
	return nil
}

func (s *Sui) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (s *Sui) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"chain_id": %d, "rpc_url": "%s", "noop_contract_address": "%s", "gas_budget": "10000000","timeout": 5000000000, "mnemonic_env": "SUI_MNEMONIC"}`,
		s.ChainID,
		s.RpcEndpoint,
		s.NoopContractAddress,
	)
}

func (s *Sui) SetRPCEndpoint(rpc string) {
	s.RpcEndpoint = rpc
}

func (s *Sui) GetLightNodeEndpoint() string {
	return ""
}

func (s *Sui) GetNetworkName() string {
	return "sui"
}

func (s *Sui) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (s *Sui) GetKeyName() string {
	return "sui"
}

func (s *Sui) GetNamespaceID() string {
	return ""
}

func (s *Sui) GetAppID() uint32 {
	return 0
}

func (s *Sui) getBalance() (*big.Int, error) {
	ctx := context.Background()
	cli := sui.NewSuiClient(s.RpcEndpoint)

	rsp, err := cli.SuiXGetBalance(ctx, models.SuiXGetBalanceRequest{
		Owner:    s.Address,
		CoinType: "0x2::sui::SUI",
	})
	if err != nil {
		return nil, err
	}

	bigIntValue := new(big.Int)

	bigIntValue, success := bigIntValue.SetString(rsp.TotalBalance, 10)
	if !success {
		return nil, fmt.Errorf("⚠️ Error converting string to big.Int")
	}

	return bigIntValue, nil
}
