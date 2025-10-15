package solana

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os/exec"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"
)

const (
	ConfigFileName = "solana.toml"
)

type Solana struct {
	Root        string
	Address     string
	RpcEndpoint string
}

func NewSolana(root string) *Solana {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	solanaConfig, err := loadConfigFromTOML(cfgPath)
	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.SolanaMainnet)
		} else {
			daNetwork = string(consts.SolanaTestnet)
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			pterm.Error.Printf("DA network configuration not found for: %s", daNetwork)
			return &solanaConfig
		}

		solanaConfig.RpcEndpoint = daData.RpcUrl
		solanaConfig.Root = root

		useExistingRpc, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to use your own RPC endpoint??",
		).Show()

		if useExistingRpc {
			solanaConfig.RpcEndpoint, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your RPC endpoint",
			).Show()
		}

		solanaConfig.Address, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"> Enter your sequencer Solana address with funds",
		).Show()

		pterm.DefaultSection.WithIndentCharacter("🔔").Println("Please fund your Solana addresses below")
		pterm.DefaultBasicText.Println(pterm.LightGreen(solanaConfig.Address))

		for {
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallet is funded",
				).Show()

			if !proceed {
				pterm.Error.Println("Solana addr needs to be funded")
				continue
			}

			balance, err := solanaConfig.getBalance()
			if err != nil {
				pterm.Println("Error getting balance:", err)
				continue
			}

			balanceBig := new(big.Int).SetUint64(balance)
			if balanceBig.Cmp(big.NewInt(0)) > 0 {
				pterm.Println("Wallet funded with balance:", balance)
				break
			}
			pterm.Error.Println("Solana wallet needs to be funded")
		}

		pterm.Warning.Print("You will need to save Solana keypath to an environment variable named SOLANA_KEYPATH")

		err = writeConfigToTOML(cfgPath, solanaConfig)
		if err != nil {
			panic(err)
		}
	}
	return &solanaConfig
}

func (s *Solana) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (s *Solana) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return &keys.KeyInfo{
		Address: s.Address,
	}, nil
}

func (s *Solana) GetRootDirectory() string {
	return s.Root
}

func (s *Solana) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	balance, err := s.getBalance()
	if err != nil {
		return nil, fmt.Errorf("failed to get DA balance: %w", err)
	}

	exp := new(big.Int).Exp(big.NewInt(10), big.NewInt(9), nil) // SOL has 9 decimals
	required := new(big.Int).Mul(big.NewInt(1), exp)
	balanceBig := new(big.Int).SetUint64(balance)
	if required.Cmp(balanceBig) > 0 {
		return []keys.NotFundedAddressData{
			{
				KeyName:         s.GetKeyName(),
				Address:         s.Address,
				CurrentBalance:  balanceBig,
				RequiredBalance: required,
				Denom:           "SOL",
				Network:         string(consts.Solana),
			},
		}, nil
	}
	return nil, nil
}

func (s *Solana) GetStartDACmd() *exec.Cmd {
	return nil
}

func (s *Solana) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	balance, err := s.getBalance()
	if err != nil {
		return nil, err
	}

	return []keys.AccountData{
		{
			Address: s.Address,
			Balance: cosmossdktypes.Coin{
				Denom:  "SOL",
				Amount: cosmossdkmath.NewIntFromUint64(balance),
			},
		},
	}, nil
}

func (s *Solana) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"endpoint":"%s","keypath_env":"SOLANA_KEYPATH","program_address":"%s"}`,
		s.RpcEndpoint,
		s.Address,
	)
}

func (s *Solana) SetRPCEndpoint(rpc string) {
	s.RpcEndpoint = rpc
}

func (s *Solana) GetLightNodeEndpoint() string {
	return ""
}

func (s *Solana) GetNetworkName() string {
	return "solana"
}

func (s *Solana) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (s *Solana) GetKeyName() string {
	return "solana"
}

func (s *Solana) GetNamespaceID() string {
	return ""
}

func (s *Solana) GetAppID() uint32 {
	return 0
}

// Struct cho RPC request
type RPCRequest struct {
	Jsonrpc string        `json:"jsonrpc"`
	ID      int           `json:"id"`
	Method  string        `json:"method"`
	Params  []interface{} `json:"params"`
}

// Struct cho RPC response
type RPCResponse struct {
	Result struct {
		Value uint64 `json:"value"`
	} `json:"result"`
	Error interface{} `json:"error"`
}

func (s *Solana) getBalance() (uint64, error) {
	reqBody := RPCRequest{
		Jsonrpc: "2.0",
		ID:      1,
		Method:  "getBalance",
		Params:  []interface{}{s.Address},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return 0, err
	}

	resp, err := http.Post(s.RpcEndpoint, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var rpcResp RPCResponse
	err = json.Unmarshal(body, &rpcResp)
	if err != nil {
		return 0, err
	}

	if rpcResp.Error != nil {
		return 0, fmt.Errorf("RPC error: %v", rpcResp.Error)
	}

	return rpcResp.Result.Value, nil
}

func (s *Solana) GetPrivateKey() (string, error) {
	return "", nil
}

func (s *Solana) SetMetricsEndpoint(endpoint string) {
	// Not implemented for Solana
}
