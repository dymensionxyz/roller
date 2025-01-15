package oracle

import (
	"context"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/utils/roller"
)

// EVMDeployer implements ContractDeployer for EVM chains
type EVMDeployer struct {
	config     *OracleConfig
	rollerData roller.RollappConfig
}

// NewEVMDeployer creates a new EVMDeployer instance
func NewEVMDeployer(rollerData roller.RollappConfig) (*EVMDeployer, error) {
	config := NewOracle(rollerData)

	if err := config.SetKey(rollerData); err != nil {
		return nil, fmt.Errorf("failed to set oracle key: %w", err)
	}

	return &EVMDeployer{
		config:     config,
		rollerData: roller.RollappConfig{},
	}, nil
}

func (e *EVMDeployer) Config() *OracleConfig {
	return e.config
}

// DownloadContract implements ContractDeployer.DownloadContract for EVM
func (e *EVMDeployer) DownloadContract(url string) error {
	contractPath := filepath.Join(e.config.ConfigDirPath, "centralized_oracle.sol")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(e.config.ConfigDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

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

// DeployContract implements ContractDeployer.DeployContract for EVM
func (e *EVMDeployer) DeployContract(
	ctx context.Context,
	privateKey *ecdsa.PrivateKey,
	contractCode []byte,
) (string, error) {
	return "", errors.New("not implemented")
	// chainID, err := e.client.ChainID(ctx)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to get chain ID: %w", err)
	// }

	// auth, err := bind.NewKeyedTransactorWithChainID(privateKey, chainID)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to create transactor: %w", err)
	// }

	// // Get the next nonce for the sender address
	// nonce, err := e.client.PendingNonceAt(ctx, auth.From)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to get nonce: %w", err)
	// }

	// // Get gas price
	// gasPrice, err := e.client.SuggestGasPrice(ctx)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to get gas price: %w", err)
	// }

	// // Set up transaction options
	// auth.Nonce = big.NewInt(int64(nonce))
	// auth.Value = big.NewInt(0) // No ETH value to send
	// auth.GasPrice = gasPrice
	// auth.GasLimit = uint64(3000000) // You might want to estimate this

	// // Create and sign the transaction
	// tx := &bind.TransactOpts{
	// 	From:     auth.From,
	// 	Nonce:    auth.Nonce,
	// 	Signer:   auth.Signer,
	// 	Value:    auth.Value,
	// 	GasPrice: auth.GasPrice,
	// 	GasLimit: auth.GasLimit,
	// 	Context:  ctx,
	// }

	// // Deploy the contract
	// address, tx, _, err := bind.DeployContract(tx, bind.NewBoundContract(common.Address{}, nil, e.client, e.client, e.client), contractCode)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to deploy contract: %w", err)
	// }

	// // Wait for the transaction to be mined
	// receipt, err := bind.WaitMined(ctx, e.client, tx)
	// if err != nil {
	// 	return "", fmt.Errorf("failed to wait for contract deployment: %w", err)
	// }

	// if receipt.Status == 0 {
	// 	return "", fmt.Errorf("contract deployment failed")
	// }

	// return address.Hex(), nil
}
