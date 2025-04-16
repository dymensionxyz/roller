package loadnetwork

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os/exec"
	"strconv"

	cosmossdkmath "cosmossdk.io/math"
	cosmossdktypes "github.com/cosmos/cosmos-sdk/types"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/errorhandling"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/pterm/pterm"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	ConfigFileName               = "load_network.toml"
	mnemonicEntropySize          = 256
	keyringNetworkID      uint16 = 42
	requiredAVL                  = 1
	DefaultTestnetChainID        = 9496
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

type LoadNetwork struct {
	Root        string
	PrivateKey  string
	RpcEndpoint string
	ChainID     uint32
}

func (w *LoadNetwork) GetPrivateKey() (string, error) {
	return w.PrivateKey, nil
}

func (w *LoadNetwork) SetMetricsEndpoint(endpoint string) {
}

func NewLoadNetwork(root string) *LoadNetwork {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	loadNetworkConfig, err := loadConfigFromTOML(cfgPath)
	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.LoadNetworkMainnet)
		} else {
			daNetwork = string(consts.LoadNetworkTestnet)
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			pterm.Error.Printf("DA network configuration not found for: %s", daNetwork)
			return &loadNetworkConfig
		}

		useExistingWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
			"would you like to import an existing LoadNetwork wallet?",
		).Show()

		if useExistingWallet {
			loadNetworkConfig.PrivateKey, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
				"> Enter your PrivateKey without 0x",
			).Show()
		} else {
			privateKey, err := crypto.GenerateKey()
			if err != nil {
				panic(err)
			}

			privateKeyBytes := crypto.FromECDSA(privateKey)
			// privateKeyHex := hex.EncodeToString(privateKeyBytes)
			// if err != nil {
			// 	panic(err)
			// }

			loadNetworkConfig.PrivateKey = fmt.Sprintf("%x", string(privateKeyBytes))

			fmt.Printf("\t%s\n", loadNetworkConfig.PrivateKey)
			fmt.Println()
			fmt.Println(pterm.LightYellow("💡 save this information and keep it safe"))
		}

		for {
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallet is funded",
				).Show()

			if !proceed {
				pterm.Error.Println("LoadNetwork wallet needs to be funded!")
				continue
			}

			balance, err := GetBalance(daData.ApiUrl, loadNetworkConfig.PrivateKey)
			if err != nil {
				pterm.Println("Error getting balance:", err)
				continue
			}

			balanceFloat, err := strconv.ParseFloat(balance, 64)
			if err != nil {
				pterm.Println("Error parsing balance:", err)
				continue
			}

			if balanceFloat > 0 {
				pterm.Println("Wallet funded with balance:", balance)
				break
			}

			pterm.Error.Println("LoadNetwork wallet needs to be funded!")
		}

		loadNetworkConfig.RpcEndpoint = daData.ApiUrl
		loadNetworkConfig.Root = root
		loadNetworkConfig.ChainID = DefaultTestnetChainID

		err = writeConfigToTOML(cfgPath, loadNetworkConfig)
		if err != nil {
			panic(err)
		}
	}
	return &loadNetworkConfig
}

func (w *LoadNetwork) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (w *LoadNetwork) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (w *LoadNetwork) GetRootDirectory() string {
	return w.Root
}

func (w *LoadNetwork) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (w *LoadNetwork) GetStartDACmd() *exec.Cmd {
	return nil
}

func (w *LoadNetwork) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
	balance, err := GetBalance(cfg.DA.ApiUrl, w.PrivateKey)
	if err != nil {
		return nil, err
	}
	address, _, err := getAddressFromPrivateKey(w.PrivateKey)
	if err != nil {
		return nil, err
	}

	balanceInt, ok := cosmossdkmath.NewIntFromString(balance)
	if !ok {
		return nil, fmt.Errorf("Can not convert from String to Int")
	}
	return []keys.AccountData{
		{
			Address: address.String(),
			Balance: cosmossdktypes.Coin{
				Denom:  consts.Denoms.LoadNetwork,
				Amount: balanceInt,
			},
		},
	}, nil
}

func (w *LoadNetwork) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"endpoint": "%s", "chain_id": %d,"private_key_hex": "%s"}`,
		w.RpcEndpoint,
		w.ChainID,
		w.PrivateKey,
	)
}

func (w *LoadNetwork) SetRPCEndpoint(rpc string) {
	w.RpcEndpoint = rpc
}

func (w *LoadNetwork) GetLightNodeEndpoint() string {
	return ""
}

func (w *LoadNetwork) GetNetworkName() string {
	return "loadnetwork"
}

func (w *LoadNetwork) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (w *LoadNetwork) GetKeyName() string {
	return "loadnetwork"
}

func (w *LoadNetwork) GetNamespaceID() string {
	return ""
}

func (w *LoadNetwork) GetAppID() uint32 {
	return 0
}

func getAddressFromPrivateKey(privKey string) (common.Address, *ecdsa.PrivateKey, error) {
	// Getting public address from private key
	pKeyBytes, err := hexutil.Decode("0x" + privKey)
	if err != nil {
		return common.Address{}, nil, err
	}
	// Convert the private key bytes to an ECDSA private key.
	ecdsaPrivateKey, err := crypto.ToECDSA(pKeyBytes)
	if err != nil {
		return common.Address{}, nil, err
	}
	// Extract the public key from the ECDSA private key.
	publicKey := ecdsaPrivateKey.Public()
	publicKeyECDSA, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return common.Address{}, nil, fmt.Errorf("error casting public key to ECDSA")
	}

	// Compute the Ethereum address of the signer from the public key.
	fromAddress := crypto.PubkeyToAddress(*publicKeyECDSA)
	return fromAddress, ecdsaPrivateKey, nil
}

func GetBalance(jsonRPCURL, key string) (string, error) {
	address, _, err := getAddressFromPrivateKey(key)
	if err != nil {
		return "", err
	}
	payload := RequestPayload{
		JSONRPC: "2.0",
		Method:  "eth_getBalance",
		Params:  []interface{}{address, "latest"},
		ID:      1,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	resp, err := http.Post(jsonRPCURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	var response EthBalanceResponse

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return "", err
	}

	balance := new(big.Int)
	balance.SetString(response.Result[2:], 16)

	ethBalance := new(big.Float).Quo(new(big.Float).SetInt(balance), big.NewFloat(1e18))

	return ethBalance.Text('f', 6), nil
}
