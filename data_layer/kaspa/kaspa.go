package kaspa

import (
	"bytes"
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
	ConfigFileName      = "kaspa.toml"
	MnemonicEntropySize = 256
	requiredKAS         = 1
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
			kaspaConfig.Network = "mainnet"
		} else {
			daNetwork = string(consts.KaspaTestnet)
			kaspaConfig.Network = "testnet"
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			pterm.Error.Printf("DA network configuration not found for: %s", daNetwork)
			return &kaspaConfig
		}

		kaspaConfig.ApiUrl = daData.ApiUrl
		kaspaConfig.Root = root

		useExistingGrpcAddress, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to use your own gRPC endpoint??",
		).Show()

		if useExistingGrpcAddress {
			kaspaConfig.GrpcAddress, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your gRPC endpoint",
			).Show()
		} else {
			kaspaConfig.GrpcAddress = "localhost:16210"
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

type GetUtxosParams struct {
	Address string `json:"address"`
}

type JsonRpcRequest struct {
	Jsonrpc string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  GetUtxosParams `json:"params"`
	ID      int            `json:"id"`
}

type UTXO struct {
	Amount uint64 `json:"amount"`
}

type JsonRpcResponse struct {
	Result struct {
		Entries []UTXO `json:"entries"`
	} `json:"result"`
	Error interface{} `json:"error"`
}

func (k *Kaspa) getBalance() (uint64, error) {
	reqBody := JsonRpcRequest{
		Jsonrpc: "2.0",
		Method:  "getUtxosByAddress",
		Params:  GetUtxosParams{Address: k.Address},
		ID:      1,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(k.ApiUrl, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	respData, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	var rpcResp JsonRpcResponse
	if err := json.Unmarshal(respData, &rpcResp); err != nil {
		return 0, err
	}

	if rpcResp.Error != nil {
		return 0, fmt.Errorf("RPC error: %v", rpcResp.Error)
	}

	var total uint64 = 0
	for _, utxo := range rpcResp.Result.Entries {
		total += utxo.Amount
	}

	return total, nil
}
