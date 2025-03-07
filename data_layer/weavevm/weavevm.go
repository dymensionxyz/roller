package weavevm

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
	ConfigFileName               = "weavevm.toml"
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

type WeaveVM struct {
	Root        string
	PrivateKey  string
	RpcEndpoint string
	ChainID     uint32
}

func (w *WeaveVM) GetPrivateKey() (string, error) {
	return w.PrivateKey, nil
}

func (w *WeaveVM) SetMetricsEndpoint(endpoint string) {
}

func NewWeaveVM(root string) *WeaveVM {
	var daNetwork string

	rollerData, err := roller.LoadConfig(root)
	errorhandling.PrettifyErrorIfExists(err)

	cfgPath := GetCfgFilePath(root)
	weavevmConfig, err := loadConfigFromTOML(cfgPath)
	if err != nil {
		if rollerData.HubData.Environment == "mainnet" {
			daNetwork = string(consts.WeaveVMMainnet)
		} else {
			daNetwork = string(consts.WeaveVMTestnet)
		}

		weavevmConfig.PrivateKey, _ = pterm.DefaultInteractiveTextInput.WithDefaultText(
			"> Enter your PrivateKey without 0x",
		).Show()

		proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
			WithDefaultText(
				"press 'y' when the wallet are funded",
			).Show()

		if !proceed {
			panic(fmt.Errorf("WeaveVM wallet need to be fund!"))
		}

		daData, exists := consts.DaNetworks[daNetwork]
		if !exists {
			panic(fmt.Errorf("DA network configuration not found for: %s", daNetwork))
		}

		balance, err := GetBalance(daData.ApiUrl, weavevmConfig.PrivateKey)
		if err != nil {
			panic(err)
		}

		balanceFloat, err := strconv.ParseFloat(balance, 64)
		if err != nil {
			panic(err)
		}

		if balanceFloat == 0 {
			panic(fmt.Errorf("WeaveVM wallet need to be fund!"))
		}

		pterm.Println("WeaveVM Balance: ", balanceFloat)

		weavevmConfig.RpcEndpoint = daData.ApiUrl
		weavevmConfig.Root = root
		weavevmConfig.ChainID = DefaultTestnetChainID

		err = writeConfigToTOML(cfgPath, weavevmConfig)
		if err != nil {
			panic(err)
		}
	}
	return &weavevmConfig
}

func (w *WeaveVM) InitializeLightNodeConfig() (string, error) {
	return "", nil
}

func (w *WeaveVM) GetDAAccountAddress() (*keys.KeyInfo, error) {
	return nil, nil
}

func (w *WeaveVM) GetRootDirectory() string {
	return w.Root
}

func (w *WeaveVM) CheckDABalance() ([]keys.NotFundedAddressData, error) {
	return nil, nil
}

func (w *WeaveVM) GetStartDACmd() *exec.Cmd {
	return nil
}

func (w *WeaveVM) GetDAAccData(cfg roller.RollappConfig) ([]keys.AccountData, error) {
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
				Denom:  consts.Denoms.WeaveVM,
				Amount: balanceInt,
			},
		},
	}, nil
}

func (w *WeaveVM) GetSequencerDAConfig(_ string) string {
	return fmt.Sprintf(
		`{"endpoint": "%s", "chain_id": %d,"private_key_hex": "%s"}`,
		w.RpcEndpoint,
		w.ChainID,
		w.PrivateKey,
	)
}

func (w *WeaveVM) SetRPCEndpoint(rpc string) {
	w.RpcEndpoint = rpc
}

func (w *WeaveVM) GetLightNodeEndpoint() string {
	return ""
}

func (w *WeaveVM) GetNetworkName() string {
	return "weavevm"
}

func (w *WeaveVM) GetStatus(c roller.RollappConfig) string {
	return "Active"
}

func (w *WeaveVM) GetKeyName() string {
	return "weavevm"
}

func (w *WeaveVM) GetNamespaceID() string {
	return ""
}

func (w *WeaveVM) GetAppID() uint32 {
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
