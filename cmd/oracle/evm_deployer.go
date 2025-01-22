package oracle

import (
	"bytes"
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cosmossdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	ethtypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
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
		rollerData: rollerData,
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
) (string, error) {
	// Compile the contract
	contractPath := filepath.Join(e.config.ConfigDirPath, "centralized_oracle.sol")
	bytecode, _, err := compileContract(contractPath)
	if err != nil {
		return "", fmt.Errorf("failed to compile contract: %w", err)
	}

	// Convert string private key to ECDSA private key
	pterm.Info.Printfln("deploying contract with private key: %s", e.config.PrivateKey)
	raResp, err := rollapp.GetMetadataFromChain(e.rollerData.RollappID, e.rollerData.HubData)
	if err != nil {
		return "", fmt.Errorf("failed to get rollapp metadata: %v", err)
	}

	var balanceDenom string
	if raResp.Rollapp.GenesisInfo.NativeDenom == nil {
		balanceDenom = consts.Denoms.HubIbcOnRollapp
	} else {
		balanceDenom = raResp.Rollapp.GenesisInfo.NativeDenom.Base
	}
	for {
		balance, err := keys.QueryBalance(
			keys.ChainQueryConfig{
				Denom:  balanceDenom,
				RPC:    "http://localhost:26657",
				Binary: consts.Executables.RollappEVM,
			}, e.config.KeyAddress,
		)
		if err != nil {
			return "", fmt.Errorf("failed to query balance: %v", err)
		}

		one, _ := cosmossdkmath.NewIntFromString("1000000000000000000")
		isAddrFunded := balance.Amount.GTE(one)

		if !isAddrFunded {
			oracleKeys, err := getOracleKeyConfig()
			if err != nil {
				return "", fmt.Errorf("failed to get oracle keys: %v", err)
			}
			kc := oracleKeys[0]

			ki, err := kc.Info(e.rollerData.Home)
			if err != nil {
				return "", fmt.Errorf("failed to get key info: %v", err)
			}

			pterm.DefaultSection.WithIndentCharacter("🔔").
				Println("Please fund the addresses below be able to deploy an oracle")
			ki.Print(keys.WithName())
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallets are funded",
				).Show()
			if !proceed {
				return "", fmt.Errorf("cancelled by user")
			}
		} else {
			break
		}
	}

	// Deploy the contract using deployEvmContract
	contractAddress, err := deployEvmContract(
		bytecode,
		e.config.PrivateKey,
		e.config.EcdsaPrivateKey,
	)
	if err != nil {
		return "", fmt.Errorf("failed to deploy contract: %w", err)
	}

	return contractAddress.Hex(), nil
}

// contract deployment code was adapted from https://github.com/bcdevtools/devd/blob/main/cmd/tx/deploy-contract.go

func deployEvmContract(
	bytecode string,
	privateKey string,
	ecdsaPrivateKey *ecdsa.PrivateKey,
) (*common.Address, error) {
	ethClient8545, _ := ethclient.Dial("http://localhost:8545")
	if ethClient8545 == nil {
		return nil, errors.New("failed to connect to local evm rpc endpoint")
	}

	// Convert the private key to hex string
	pterm.Warning.Println("private key received:" + privateKey)
	pterm.Warning.Println("ecdsa private key received:" + ecdsaPrivateKey.D.String())

	pterm.Warning.Println("public key received:" + ecdsaPrivateKey.PublicKey.X.String())
	publicKey := ecdsaPrivateKey.Public()
	ecdsaPubKey, ok := publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("failed to convert secret public key to ECDSA")
	}

	// Convert ECDSA public key to compressed pubkey bytes
	pubKeyBytes := crypto.CompressPubkey(ecdsaPubKey)

	// Create Cosmos SDK public key
	var pubKey cryptotypes.PubKey = &secp256k1.PubKey{Key: pubKeyBytes}

	// Get Cosmos SDK address
	fromAddress := sdk.AccAddress(pubKey.Address())

	// Convert to Ethereum address format
	ethAddr := common.BytesToAddress(fromAddress.Bytes())

	nonce, err := ethClient8545.NonceAt(context.Background(), ethAddr, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	chainId, err := ethClient8545.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	bytecode = strings.TrimSuffix(bytecode, "0x")

	deploymentBytes, err := hex.DecodeString(bytecode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment bytecode: %w", err)
	}

	balance, err := ethClient8545.BalanceAt(context.Background(), ethAddr, nil)
	if err != nil {
		fmt.Printf("Error getting balance: %v\n", err)
	} else {
		fmt.Printf("Balance: %s wei\n", balance.String())
	}

	txData := ethtypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(20_000_000_000),
		Gas:      4_000_000,
		To:       nil,
		Data:     deploymentBytes,
		Value:    common.Big0,
	}
	tx := ethtypes.NewTx(&txData)

	newContractAddress := crypto.CreateAddress(ethAddr, nonce)

	fmt.Println("Deploying new contract using account", ethAddr)

	signedTx, err := ethtypes.SignTx(tx, ethtypes.LatestSignerForChainID(chainId), ecdsaPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	var buf bytes.Buffer
	err = signedTx.EncodeRLP(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx: %w", err)
	}

	fmt.Println("Tx hash", signedTx.Hash())

	err = ethClient8545.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send tx: %w", err)
	}

	if tx := waitForEthTx(ethClient8545, signedTx.Hash()); tx != nil {
		fmt.Println("New contract deployed at:")
	} else {
		fmt.Println("Timed-out waiting for tx to be mined, contract may have been deployed.")
		fmt.Println("Expected contract address:")
	}
	fmt.Println(newContractAddress)

	return &newContractAddress, nil
}

func waitForEthTx(ethClient8545 *ethclient.Client, txHash common.Hash) *ethtypes.Transaction {
	for try := 1; try <= 6; try++ {
		tx, _, err := ethClient8545.TransactionByHash(context.Background(), txHash)
		if err == nil && tx != nil {
			return tx
		}

		time.Sleep(time.Second)
	}

	return nil
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
