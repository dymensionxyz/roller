package oracleutils

import (
	"context"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/pterm/pterm"
	"gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/utils/roller"
)

// WasmDeployer implements ContractDeployer for WASM chains
type WasmDeployer struct {
	config     *OracleConfig
	rollerData roller.RollappConfig
	KeyData    struct {
		KeyData
		PrivateKey string
	}
}

// NewWasmDeployer creates a new WasmDeployer instance
func NewWasmDeployer(rollerData roller.RollappConfig) (*WasmDeployer, error) {
	config := NewOracleConfig(rollerData)
	d := &WasmDeployer{
		config:     config,
		rollerData: rollerData,
	}

	err := d.SetKey()
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (w *WasmDeployer) PrivateKey() string {
	return w.KeyData.PrivateKey
}

func (w *WasmDeployer) ClientConfigPath() string {
	return filepath.Join(w.config.ConfigDirPath, "config.yaml")
}

func (w *WasmDeployer) IsContractDeployed() (string, bool) {
	configDir := filepath.Dir(w.config.ConfigDirPath)
	configFilePath := w.ClientConfigPath()

	if _, err := os.Stat(configDir); err == nil {
		if _, err := os.Stat(configFilePath); err == nil {
			configData, err := os.ReadFile(configFilePath)
			if err != nil {
				return "", false
			}

			var config struct {
				ChainClient struct {
					OracleContractAddress string `yaml:"oracleContractAddress"`
				} `yaml:"chainClient"`
			}

			if err := yaml.Unmarshal(configData, &config); err != nil {
				return "", false
			}

			if config.ChainClient.OracleContractAddress != "" {
				w.config.ContractAddress = config.ChainClient.OracleContractAddress
				return w.config.ContractAddress, true
			}
		}
	}

	return "", false
}

func (w *WasmDeployer) SetKey() error {
	addr, isExisting, err := generateRaOracleKeys(w.rollerData.Home, w.rollerData)
	if err != nil {
		return fmt.Errorf("failed to retrieve oracle keys: %v", err)
	}

	if len(addr) == 0 {
		return fmt.Errorf("no oracle keys generated")
	}

	w.KeyData.Address = addr[0].Address
	w.KeyData.Name = addr[0].Name

	if !isExisting {
		hexKey, err := GetSecp256k1PrivateKey(addr[0].Mnemonic)
		if err != nil {
			return err
		}

		w.KeyData.PrivateKey = hexKey
	} else {
		w.KeyData.PrivateKey = ""
	}

	return nil
}

func (w *WasmDeployer) Config() *OracleConfig {
	return w.config
}

// DownloadContract implements ContractDeployer.DownloadContract for WASM
func (w *WasmDeployer) DownloadContract(url string, outputName string, oracleType string) error {
	configDirPath := filepath.Join(w.config.ConfigDirPath, oracleType)
	contractPath := filepath.Join(configDirPath, outputName)

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(configDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Download the contract
	// nolint: gosec
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download contract: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download contract, status: %s", resp.Status)
	}

	// Read the contract bytes
	contractCode, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read contract code: %w", err)
	}

	// Save the contract file
	if err := os.WriteFile(contractPath, contractCode, 0o644); err != nil {
		return fmt.Errorf("failed to save contract: %w", err)
	}

	pterm.Info.Println("contract downloaded successfully to " + contractPath)

	return nil
}

// DeployContract implements ContractDeployer.DeployContract for WASM
func (w *WasmDeployer) DeployContract(
	ctx context.Context,
	contractName string,
) (string, error) {
	// Store the contract
	if err := w.config.StoreWasmContract(w.rollerData, contractName); err != nil {
		return "", fmt.Errorf("failed to store contract: %w", err)
	}

	// Wait for the transaction to be processed
	time.Sleep(time.Second * 2)

	// Get the code ID
	codeID, err := w.config.GetCodeID(contractName)
	if err != nil {
		return "", fmt.Errorf("failed to get code ID: %w", err)
	}

	if codeID == "" {
		return "", fmt.Errorf("failed to get code ID after storing contract")
	}

	w.config.CodeID = codeID

	// Check for existing contracts first
	contracts, err := w.config.ListContracts(w.rollerData)
	if err != nil {
		return "", fmt.Errorf("failed to list contracts: %w", err)
	}

	// If contract already exists, return its address
	if len(contracts) > 0 {
		w.config.ContractAddress = contracts[0]
		return contracts[0], nil
	}

	// Instantiate a new contract
	if err := w.config.InstantiateContract(w.rollerData); err != nil {
		return "", fmt.Errorf("failed to instantiate contract: %w", err)
	}

	// Get the newly created contract address
	contracts, err = w.config.ListContracts(w.rollerData)
	if err != nil {
		return "", fmt.Errorf("failed to get contract address: %w", err)
	}

	if len(contracts) == 0 {
		return "", fmt.Errorf("no contract address found after deployment")
	}

	w.config.ContractAddress = contracts[0]
	return contracts[0], nil
}

func GetSecp256k1PrivateKey(mnemonic string) (string, error) {
	if !bip39.IsMnemonicValid(mnemonic) {
		return "", fmt.Errorf("invalid mnemonic")
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return "", fmt.Errorf("failed to generate seed: %w", err)
	}

	hdPath := "m/44'/60'/0'/0/0"
	master, ch := hd.ComputeMastersFromSeed(seed)
	privKey, err := hd.DerivePrivateKeyForPath(master, ch, hdPath)
	if err != nil {
		return "", fmt.Errorf("failed to derive private key: %w", err)
	}

	return hex.EncodeToString(privKey), nil
}
