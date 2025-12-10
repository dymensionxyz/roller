package kaspa

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os/exec"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"
)

const (
	ConfigFileName          = "kaspa.toml"
	MnemonicEntropySize     = 256
	requiredKAS             = 1
	defaultKaspaTestnetGRPC = "api-kaspa.mzonder.com:16210"
	defaultKaspaMainnetGRPC = "91.84.65.9:16210"
	MnemonicEnvVar          = "KASPA_MNEMONIC"
)

type Kaspa struct {
	Root        string
	Address     string
	GrpcAddress string
	Network     string
	ApiUrl      string
	Mnemonic    string
}

func (k *Kaspa) GetPrivateKey() (string, error) {
	return k.Address, nil
}

func (k *Kaspa) SetMetricsEndpoint(endpoint string) {
}

func NewKaspa(root string) *Kaspa {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	kaspaConfig, err := loadConfigFromTOML(cfgPath)
	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.KaspaMainnet)
			kaspaConfig.Network = "kaspa-mainnet"
			kaspaConfig.GrpcAddress = defaultKaspaMainnetGRPC
		} else {
			daNetwork = string(consts.KaspaTestnet)
			kaspaConfig.Network = "kaspa-testnet-10"
			kaspaConfig.GrpcAddress = defaultKaspaTestnetGRPC
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			pterm.Error.Printf("DA network configuration not found for: %s", daNetwork)
			return &kaspaConfig
		}

		kaspaConfig.ApiUrl = daData.ApiUrl
		kaspaConfig.Root = root

		useExistingApiEndpoint, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to use your own API endpoint?",
		).Show()

		if useExistingApiEndpoint {
			kaspaConfig.ApiUrl, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your API endpoint",
			).Show()
		}

		useExistingGrpcAddress, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to use your own gRPC endpoint??",
		).Show()

		if useExistingGrpcAddress {
			kaspaConfig.GrpcAddress, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your gRPC endpoint",
			).Show()
		}

		kaspaConfig.Address, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"> Enter your Kaspa Address",
		).Show()

		pterm.DefaultSection.WithIndentCharacter("🔔").Println("Please fund your Kaspa addresses below")
		pterm.DefaultBasicText.Println(pterm.LightGreen(kaspaConfig.Address))

		for {
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallet is funded",
				).Show()

			if !proceed {
				pterm.Error.Println("Kaspa addr needs to be funded")
				continue
			}

			balance, err := kaspaConfig.getBalance()
			if err != nil {
				pterm.Println("Error getting balance:", err)
				continue
			}

			balanceBig := new(big.Int).SetUint64(balance)
			if balanceBig.Cmp(big.NewInt(0)) > 0 {
				pterm.Println("Wallet funded with balance:", balance)
				break
			}
			pterm.Error.Println("Kaspa wallet needs to be funded")
		}

		err = writeConfigToTOML(cfgPath, kaspaConfig)
		if err != nil {
			panic(err)
		}
	}
	return &kaspaConfig
}

func (k *Kaspa) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (k *Kaspa) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return &keys.KeyInfo{
		Address: k.Address,
	}, nil
}

func (k *Kaspa) GetRootDirectory() string {
	return k.Root
}

func (k *Kaspa) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	balance, err := k.getBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get DA balance: %w", err)
	}

	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(8), nil) // Kaspa has 8 decimals
	required := new(big.Int).Mul(big.NewInt(requiredKAS), exp)
	balanceBig := new(big.Int).SetUint64(balance)
	if required.Cmp(balanceBig) > 0 {
		return []keys.NotFundedAddressData{
			{
				KeyName:         k.GetKeyName(),
				Address:         k.Address,
				CurrentBalance:  balanceBig,
				RequiredBalance: required,
				Denom:           "KAS",
				Network:         string(consts.Kaspa),
			},
		}, nil
	}
	return nil, nil
}

func (k *Kaspa) GetStartDACmd() *exec.Cmd {
	return nil
}

func (k *Kaspa) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	return nil, nil
}

func (k *Kaspa) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"api_url":"%s","grpc_address":"%s","network":"%s","address":"%s","mnemonic_env":"KASPA_MNEMONIC"}`,
		k.ApiUrl,
		k.GrpcAddress,
		k.Network,
		k.Address,
	)
}

func (k *Kaspa) SetRPCEndpoint(rpc string) {
}

func (k *Kaspa) GetLightNodeEndpoint() string {
	return ""
}

func (k *Kaspa) GetNetworkName() string {
	return "kaspa"
}

func (k *Kaspa) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (k *Kaspa) GetKeyName() string {
	return "kaspa"
}

func (k *Kaspa) GetNamespaceID() string {
	return ""
}

func (k *Kaspa) GetAppID() uint32 {
	return 0
}

type KaspaBalanceResponse struct {
	Address string `json:"address"`
	Balance uint64 `json:"balance"` // Đơn vị: sompi
}

func (k *Kaspa) getBalance() (uint64, error) {
	url := fmt.Sprintf("%s/addresses/%s/balance", k.ApiUrl, k.Address)

	resp, err := http.Get(url)
	if err != nil {
		return 0, fmt.Errorf("failed to call Kaspa API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return 0, fmt.Errorf("Kaspa API returned status: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, fmt.Errorf("failed to read response body: %w", err)
	}

	var data KaspaBalanceResponse
	err = json.Unmarshal(body, &data)
	if err != nil {
		return 0, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return data.Balance, nil
}
