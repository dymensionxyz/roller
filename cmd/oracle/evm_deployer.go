package oracle

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies"
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
	// Compile the contract
	contractPath := filepath.Join(e.config.ConfigDirPath, "centralized_oracle.sol")
	bytecode, _, err := compileContract(contractPath)
	if err != nil {
		return "", fmt.Errorf("failed to compile contract: %w", err)
	}

	// Deploy contract using rollappd tx evm raw
	deployCmd := exec.Command(
		consts.Executables.RollappEVM,
		"tx",
		"evm",
		"raw",
		bytecode,
		"--from", consts.KeysIds.Oracle,
		"--chain-id", e.rollerData.RollappID,
		"-y",
		"--output", "json",
		"--home", e.config.ConfigDirPath,
	)

	deployOutput, err := bash.ExecCommandWithStdout(deployCmd)
	if err != nil {
		return "", fmt.Errorf("failed to deploy contract: %w", err)
	}

	var txResult struct {
		TxHash string `json:"txhash"`
	}
	if err := json.Unmarshal(deployOutput.Bytes(), &txResult); err != nil {
		return "", fmt.Errorf("failed to parse deployment result: %w", err)
	}

	// Wait a bit for the transaction to be mined
	time.Sleep(5 * time.Second)

	// Query the transaction to get the contract address
	queryCmd := exec.Command("rollappd", "query", "tx", txResult.TxHash, "-o", "json")
	queryOutput, err := bash.ExecCommandWithStdout(queryCmd)
	if err != nil {
		return "", fmt.Errorf("failed to query transaction: %w", err)
	}

	var txReceipt struct {
		Logs []struct {
			Events []struct {
				Attributes []struct {
					Key   string `json:"key"`
					Value string `json:"value"`
				} `json:"attributes"`
			} `json:"events"`
		} `json:"logs"`
	}
	if err := json.Unmarshal(queryOutput.Bytes(), &txReceipt); err != nil {
		return "", fmt.Errorf("failed to parse transaction receipt: %w", err)
	}

	// Find contract address in the logs
	var contractAddress string
	for _, log := range txReceipt.Logs {
		for _, event := range log.Events {
			for _, attr := range event.Attributes {
				if attr.Key == "contract_address" {
					contractAddress = attr.Value
					break
				}
			}
		}
	}

	if contractAddress == "" {
		return "", fmt.Errorf("contract address not found in transaction receipt")
	}

	return contractAddress, nil
}

func compileContract(contractPath string) (string, string, error) {
	// Ensure solc is installed
	if err := dependencies.InstallSolidityDependencies(); err != nil {
		return "", "", fmt.Errorf("failed to install solidity compiler: %w", err)
	}

	// Create build directory
	buildDir := filepath.Join(filepath.Dir(contractPath), "build")
	if err := os.MkdirAll(buildDir, 0o755); err != nil {
		return "", "", fmt.Errorf("failed to create build directory: %w", err)
	}

	solcPath := filepath.Join(consts.InternalBinsDir, "solc")

	// Compile contract to get bytecode
	cmd := exec.Command(solcPath, "--bin", contractPath, "-o", buildDir)
	if _, err := bash.ExecCommandWithStdout(cmd); err != nil {
		return "", "", fmt.Errorf("failed to compile contract (bytecode): %w", err)
	}

	// Compile contract to get ABI
	cmd = exec.Command(solcPath, "--abi", contractPath, "-o", buildDir)
	if _, err := bash.ExecCommandWithStdout(cmd); err != nil {
		return "", "", fmt.Errorf("failed to compile contract (ABI): %w", err)
	}

	contractName := "PriceOracle"

	binPath := filepath.Join(buildDir, fmt.Sprintf("%s.bin", contractName))
	bytecode, err := os.ReadFile(binPath)
	if err != nil {
		return "", "", fmt.Errorf("failed to read bytecode: %w", err)
	}

	runtimeBytecode := bytecode

	// nolint: errcheck
	defer os.RemoveAll(buildDir)

	return string(bytecode), string(runtimeBytecode), nil
}
