package oracle

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/utils/roller"
)

// WasmDeployer implements ContractDeployer for WASM chains
type WasmDeployer struct {
	config     *OracleConfig
	rollerData roller.RollappConfig
}

// NewWasmDeployer creates a new WasmDeployer instance
func NewWasmDeployer(rollerData roller.RollappConfig) (*WasmDeployer, error) {
	config := NewOracle(rollerData)

	if err := config.SetKey(rollerData); err != nil {
		return nil, fmt.Errorf("failed to set oracle key: %w", err)
	}

	return &WasmDeployer{
		config:     config,
		rollerData: rollerData,
	}, nil
}

func (w *WasmDeployer) Config() *OracleConfig {
	return w.config
}

// DownloadContract implements ContractDeployer.DownloadContract for WASM
func (w *WasmDeployer) DownloadContract(url string) error {
	contractPath := filepath.Join(w.config.ConfigDirPath, "centralized_oracle.wasm")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(w.config.ConfigDirPath, 0o755); err != nil {
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
	privateKey *ecdsa.PrivateKey,
	contractCode []byte,
) (string, error) {
	// Store the contract
	if err := w.config.StoreWasmContract(w.rollerData); err != nil {
		return "", fmt.Errorf("failed to store contract: %w", err)
	}

	// Wait for the transaction to be processed
	time.Sleep(time.Second * 2)

	// Get the code ID
	codeID, err := w.config.GetCodeID()
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
