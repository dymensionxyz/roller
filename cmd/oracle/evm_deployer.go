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
	"github.com/btcsuite/btcd/btcutil/hdkeychain"
	"github.com/btcsuite/btcd/chaincfg"
	cosmoshd "github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/go-bip39"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	goethcommon "github.com/ethereum/go-ethereum/common"
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
	KeyData    struct {
		KeyData
		PrivateKey *ecdsa.PrivateKey
	}
}

// NewEVMDeployer creates a new EVMDeployer instance
func NewEVMDeployer(rollerData roller.RollappConfig) (*EVMDeployer, error) {
	config := NewOracleConfig(rollerData)
	d := &EVMDeployer{
		config:     config,
		rollerData: rollerData,
	}

	err := d.SetKey()
	if err != nil {
		return nil, err
	}

	return d, nil
}

func (e *EVMDeployer) PrivateKey() string {
	return hex.EncodeToString(e.KeyData.PrivateKey.D.Bytes())
}

func (e *EVMDeployer) SetKey() error {
	addr, err := generateRaOracleKeys(e.rollerData.Home, e.rollerData)
	if err != nil {
		return fmt.Errorf("failed to retrieve oracle keys: %v", err)
	}

	if len(addr) == 0 {
		return fmt.Errorf("no oracle keys generated")
	}

	ecdsaPrivKey, err := GetEcdsaPrivateKey(addr[0].Mnemonic)
	if err != nil {
		return err
	}

	e.KeyData.Address = addr[0].Address
	e.KeyData.Name = addr[0].Name
	e.KeyData.PrivateKey = ecdsaPrivKey
	return nil
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

	raResp, err := rollapp.GetMetadataFromChain(e.rollerData.RollappID, e.rollerData.HubData)
	if err != nil {
		return "", fmt.Errorf("failed to get rollapp metadata: %v", err)
	}

	err = ensureBalance(raResp, e)
	if err != nil {
		return "", err
	}

	contractAddress, err := deployEvmContract(
		bytecode,
		e.KeyData.PrivateKey,
	)
	if err != nil {
		return "", fmt.Errorf("failed to deploy contract: %w", err)
	}

	return contractAddress.Hex(), nil
}

func ensureBalance(raResp *rollapp.ShowRollappResponse, e *EVMDeployer) error {
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
			}, e.KeyData.Address,
		)
		if err != nil {
			return fmt.Errorf("failed to query balance: %v", err)
		}

		one, _ := cosmossdkmath.NewIntFromString("1000000000000000000")
		isAddrFunded := balance.Amount.GTE(one)

		if !isAddrFunded {
			oracleKeys, err := getOracleKeyConfig(e.rollerData.RollappVMType)
			if err != nil {
				return fmt.Errorf("failed to get oracle keys: %v", err)
			}
			kc := oracleKeys[0]

			ki, err := kc.Info(e.rollerData.Home)
			if err != nil {
				return fmt.Errorf("failed to get key info: %v", err)
			}

			pterm.DefaultSection.WithIndentCharacter("🔔").
				Println("Please fund the addresses below be able to deploy an oracle")
			ki.Print(keys.WithName())
			proceed, _ := pterm.DefaultInteractiveConfirm.WithDefaultValue(false).
				WithDefaultText(
					"press 'y' when the wallets are funded",
				).Show()
			if !proceed {
				return fmt.Errorf("cancelled by user")
			}
		} else {
			break
		}
	}

	return nil
}

// contract deployment code was adapted from https://github.com/bcdevtools/devd/blob/main/cmd/tx/deploy-contract.go

func deployEvmContract(
	bytecode string,
	ecdsaPrivateKey *ecdsa.PrivateKey,
) (*goethcommon.Address, error) {
	ethClient8545, err := ethclient.Dial("http://127.0.0.1:8545")
	if err != nil {
		return nil, fmt.Errorf("failed to dial eth client: %w", err)
	}

	ecdsaPrivateKey, _, from, err := mustSecretEvmAccount(ecdsaPrivateKey)
	if err != nil {
		return nil, err
	}

	nonce, err := ethClient8545.NonceAt(context.Background(), *from, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce of sender: %w", err)
	}

	chainId, err := ethClient8545.ChainID(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	bytecode = strings.TrimPrefix(bytecode, "0x")
	deploymentBytes, err := hex.DecodeString(bytecode)
	if err != nil {
		return nil, fmt.Errorf("failed to parse deployment bytecode: %w", err)
	}

	txData := ethtypes.LegacyTx{
		Nonce:    nonce,
		GasPrice: big.NewInt(20_000_000_000),
		Gas:      400_000_000,
		To:       nil,
		Data:     deploymentBytes,
		Value:    goethcommon.Big0,
	}
	tx := ethtypes.NewTx(&txData)

	newContractAddress := crypto.CreateAddress(*from, nonce)

	fmt.Println("Deploying new contract using account", from)

	signedTx, err := ethtypes.SignTx(tx, ethtypes.NewEIP155Signer(chainId), ecdsaPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign tx: %w", err)
	}

	var buf bytes.Buffer
	err = signedTx.EncodeRLP(&buf)
	if err != nil {
		return nil, fmt.Errorf("failed to encode tx: %w", err)
	}

	fmt.Printf("Tx hash %s\n", signedTx.Hash().Hex())
	rawTxRLPHex := hex.EncodeToString(buf.Bytes())
	rawTxFile := filepath.Join("raw_tx.hex")
	if err := os.WriteFile(rawTxFile, []byte("0x"+rawTxRLPHex), 0o644); err != nil {
		return nil, fmt.Errorf("failed to write raw tx to file: %w", err)
	}
	fmt.Printf("RawTx written to: %s\n", rawTxFile)

	err = ethClient8545.SendTransaction(context.Background(), signedTx)
	if err != nil {
		return nil, fmt.Errorf("failed to send tx: %w", err)
	}

	if tx := waitForEthTx(ethClient8545, signedTx.Hash()); tx != nil {
		fmt.Printf("Contract deployed successfully at: %s\n", newContractAddress.Hex())
		return &newContractAddress, nil
	}

	return nil, fmt.Errorf("contract deployment failed - transaction was not successful")
}

func waitForEthTx(ethClient8545 *ethclient.Client, txHash common.Hash) *ethtypes.Transaction {
	for try := 1; try <= 6; try++ {
		// Get transaction receipt instead of just the transaction
		receipt, err := ethClient8545.TransactionReceipt(context.Background(), txHash)
		if err == nil && receipt != nil {
			// Check if transaction was successful
			if receipt.Status != ethtypes.ReceiptStatusSuccessful {
				fmt.Printf("Transaction failed with status: %d\n", receipt.Status)

				// Get transaction details for debugging
				tx, isPending, _ := ethClient8545.TransactionByHash(context.Background(), txHash)
				if tx != nil {
					fmt.Printf("Transaction details:\n")
					fmt.Printf("  Gas Limit: %d\n", tx.Gas())
					fmt.Printf("  Gas Price: %s\n", tx.GasPrice().String())
					fmt.Printf("  Nonce: %d\n", tx.Nonce())
					fmt.Printf("  Value: %s\n", tx.Value().String())
					fmt.Printf("  Is Pending: %v\n", isPending)
				}

				// Get latest block for gas info
				header, _ := ethClient8545.HeaderByNumber(context.Background(), nil)
				if header != nil {
					fmt.Printf("Current block gas limit: %d\n", header.GasLimit)
				}

				return nil
			}

			// For contract creation, verify code exists
			if receipt.ContractAddress != (common.Address{}) {
				code, err := ethClient8545.CodeAt(
					context.Background(),
					receipt.ContractAddress,
					nil,
				)
				if err != nil || len(code) == 0 {
					fmt.Printf(
						"No contract code found at deployed address: %s\n",
						receipt.ContractAddress.Hex(),
					)
					fmt.Printf("Gas used: %d\n", receipt.GasUsed)
					return nil
				}
				fmt.Printf("Contract code size: %d bytes\n", len(code))
			}

			// Get the transaction details
			tx, _, _ := ethClient8545.TransactionByHash(context.Background(), txHash)
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

	return string(bytecode), string(runtimeBytecode), nil
}

func GetEcdsaPrivateKey(mnemonic string) (*ecdsa.PrivateKey, error) {
	hdPathStr := cosmoshd.CreateHDPath(60, 0, 0).String()
	hdPath, err := accounts.ParseDerivationPath(hdPathStr)
	if err != nil {
		return nil, err
	}

	seed, err := bip39.NewSeedWithErrorChecking(mnemonic, "")
	if err != nil {
		return nil, err
	}

	// create a BTC-utils hd-derivation key chain
	masterKey, err := hdkeychain.NewMaster(seed, &chaincfg.MainNetParams)
	if err != nil {
		return nil, err
	}

	key := masterKey
	for _, n := range hdPath {
		key, err = key.Derive(n)
		if err != nil {
			return nil, err
		}
	}

	privateKey, err := key.ECPrivKey()
	if err != nil {
		return nil, err
	}

	// cast private key to a convertible form (single scalar field element of secp256k1)
	// and then load into ethcrypto private key format.
	return privateKey.ToECDSA(), nil
}

func mustSecretEvmAccount(
	pk *ecdsa.PrivateKey,
) (ecdsaPrivateKey *ecdsa.PrivateKey, ecdsaPubKey *ecdsa.PublicKey, account *common.Address, err error) {
	var inputSource string
	var ok bool

	ecdsaPrivateKey = pk

	publicKey := ecdsaPrivateKey.Public()
	ecdsaPubKey, ok = publicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, nil, nil, errors.New("failed to cast public key to ecdsa")
	}

	fromAddress := crypto.PubkeyToAddress(*ecdsaPubKey)
	account = &fromAddress

	fmt.Println("Account Address:", account.Hex(), "(from", inputSource, ")")

	return
}
