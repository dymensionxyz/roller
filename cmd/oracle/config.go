package oracle

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	cosmossdkmath "cosmossdk.io/math"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/tx/tx_utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

type OracleConfig struct {
	OracleVmType    string
	ConfigDirPath   string
	CodeID          string
	KeyName         string
	KeyAddress      string
	ContractAddress string
}

type KeyData struct {
	Name    string
	Address string
}

func NewOracleConfig(rollerData roller.RollappConfig) *OracleConfig {
	cd := filepath.Join(rollerData.Home, consts.ConfigDirName.Oracle)
	ovt := rollerData.RollappVMType
	return &OracleConfig{
		ConfigDirPath: cd,
		OracleVmType:  ovt.String(),
	}
}

func (o *OracleConfig) ConfigDir(rollerData roller.RollappConfig) string {
	cd := filepath.Join(rollerData.Home, consts.ConfigDirName.Oracle)
	o.ConfigDirPath = cd

	return o.ConfigDirPath
}

// generateRaOracleKeys generates a new key or retrieves an existing one from the keyring
// as the second return value it returns whether the key was generated or retrieved where
// true is retrieved and false is new generated key
func generateRaOracleKeys(
	home string,
	rollerData roller.RollappConfig,
) ([]keys.KeyInfo, bool, error) {
	oracleKeys, err := getOracleKeyConfig(rollerData.RollappVMType)
	if err != nil {
		return nil, false, err
	}

	kc := oracleKeys[0]
	ok, err := kc.IsInKeyring(home)
	if err != nil {
		return nil, false, err
	}

	if ok {
		pterm.Info.Printfln("existing oracle key found, using it")
		ki, err := kc.Info(home)
		ki.Mnemonic = "not available for already existing keys"
		if err != nil {
			return nil, false, err
		}
		return []keys.KeyInfo{*ki}, true, nil
	}

	shouldImportWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"would you like to import an existing Oracle key?",
	).Show()

	var addr []keys.KeyInfo

	if shouldImportWallet {
		ki, err := kc.Create(home)
		if err != nil {
			return nil, false, err
		}
		addr = append(addr, *ki)
	} else {
		addr, err = createOraclesKeys(rollerData)
		if err != nil {
			return nil, false, err
		}
	}

	return addr, false, nil
}

func createOraclesKeys(rollerData roller.RollappConfig) ([]keys.KeyInfo, error) {
	oracleKeys, err := getOracleKeyConfig(rollerData.RollappVMType)
	if err != nil {
		return nil, err
	}

	addresses := make([]keys.KeyInfo, 0)

	for _, key := range oracleKeys {
		var address *keys.KeyInfo
		var err error
		address, err = key.Create(rollerData.Home)
		if err != nil {
			return nil, err
		}
		addresses = append(
			addresses, keys.KeyInfo{
				Address:  address.Address,
				Name:     key.ID,
				Mnemonic: address.Mnemonic,
			},
		)

		address.Print(keys.WithName(), keys.WithMnemonic())
	}
	return addresses, nil
}

func (o *OracleConfig) StoreWasmContract(rollerData roller.RollappConfig) error {
	var cmd *exec.Cmd

	var balanceDenom string
	raResp, err := rollapp.GetMetadataFromChain(rollerData.RollappID, rollerData.HubData)
	if err != nil {
		return fmt.Errorf("failed to get rollapp metadata: %v", err)
	}

	if raResp.Rollapp.GenesisInfo.NativeDenom == nil {
		balanceDenom = consts.Denoms.HubIbcOnRollapp
	} else {
		balanceDenom = raResp.Rollapp.GenesisInfo.NativeDenom.Base
	}

	switch rollerData.RollappVMType {
	case consts.WASM_ROLLAPP:
		cmd = exec.Command(
			consts.Executables.RollappEVM,
			"tx", "wasm", "store",
			filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm"),
			"--instantiate-anyof-addresses", o.KeyAddress,
			"--from", o.KeyName,
			"--gas", "auto",
			"--gas-adjustment", "1.3",
			"--fees", fmt.Sprintf("40000000000000000%s", balanceDenom),
			"--keyring-backend", consts.SupportedKeyringBackends.Test.String(),
			"--chain-id", rollerData.RollappID,
			"--broadcast-mode", "sync",
			"--home", o.ConfigDirPath,
			"-y",
		)
	default:
		return fmt.Errorf("unsupported rollapp type: %s", rollerData.RollappVMType)
	}

	fmt.Println(cmd.String())

	for {
		balance, err := keys.QueryBalance(
			keys.ChainQueryConfig{
				Denom:  balanceDenom,
				RPC:    "http://localhost:26657",
				Binary: consts.Executables.RollappEVM,
			}, o.KeyAddress,
		)
		if err != nil {
			return fmt.Errorf("failed to query balance: %v", err)
		}

		one, _ := cosmossdkmath.NewIntFromString("1000000000000000000")
		isAddrFunded := balance.Amount.GTE(one)

		if !isAddrFunded {
			oracleKeys, err := getOracleKeyConfig(rollerData.RollappVMType)
			if err != nil {
				return fmt.Errorf("failed to get oracle keys: %v", err)
			}
			kc := oracleKeys[0]

			ki, err := kc.Info(rollerData.Home)
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

	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return fmt.Errorf("failed to store contract: %v, output: %s", err, output)
	}

	tob := bytes.NewBufferString(output.String())
	err = tx_utils.CheckTxYamlStdOut(*tob)
	if err != nil {
		return err
	}

	// Extract transaction hash
	txHash, err := bash.ExtractTxHash(output.String())
	if err != nil {
		return fmt.Errorf("failed to extract transaction hash: %v", err)
	}

	pterm.Info.Printfln("transaction hash: %s", txHash)

	// // Monitor transaction
	// wsURL := "http://localhost:26657"
	// if err := tx.MonitorTransaction(wsURL, txHash); err != nil {
	// 	return fmt.Errorf("failed to monitor transaction: %v", err)
	// }

	return nil
}

func (o *OracleConfig) StoreEvmContract(rollerData roller.RollappConfig) error {
	var cmd *exec.Cmd

	var balanceDenom string
	raResp, err := rollapp.GetMetadataFromChain(rollerData.RollappID, rollerData.HubData)
	if err != nil {
		return fmt.Errorf("failed to get rollapp metadata: %v", err)
	}

	if raResp.Rollapp.GenesisInfo.NativeDenom == nil {
		balanceDenom = consts.Denoms.HubIbcOnRollapp
	} else {
		balanceDenom = raResp.Rollapp.GenesisInfo.NativeDenom.Base
	}

	switch rollerData.RollappVMType {
	case consts.WASM_ROLLAPP:
		cmd = exec.Command(
			consts.Executables.RollappEVM,
			"tx", "wasm", "store",
			filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm"),
			"--instantiate-anyof-addresses", o.KeyAddress,
			"--from", o.KeyName,
			"--gas", "auto",
			"--gas-adjustment", "1.3",
			"--fees", fmt.Sprintf("40000000000000000%s", balanceDenom),
			"--keyring-backend", consts.SupportedKeyringBackends.Test.String(),
			"--chain-id", rollerData.RollappID,
			"--broadcast-mode", "sync",
			"--home", o.ConfigDirPath,
			"-y",
		)
	default:
		return fmt.Errorf("unsupported rollapp type: %s", rollerData.RollappVMType)
	}

	fmt.Println(cmd.String())

	for {
		balance, err := keys.QueryBalance(
			keys.ChainQueryConfig{
				Denom:  balanceDenom,
				RPC:    "http://localhost:26657",
				Binary: consts.Executables.RollappEVM,
			}, o.KeyAddress,
		)
		if err != nil {
			return fmt.Errorf("failed to query balance: %v", err)
		}

		one, _ := cosmossdkmath.NewIntFromString("1000000000000000000")
		isAddrFunded := balance.Amount.GTE(one)

		if !isAddrFunded {
			oracleKeys, err := getOracleKeyConfig(rollerData.RollappVMType)
			if err != nil {
				return fmt.Errorf("failed to get oracle keys: %v", err)
			}

			kc := oracleKeys[0]
			ki, err := kc.Info(rollerData.Home)
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

	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return fmt.Errorf("failed to store contract: %v, output: %s", err, output)
	}

	time.Sleep(time.Second * 2)

	txHash, err := bash.ExtractTxHash(output.String())
	if err != nil {
		return fmt.Errorf("failed to extract transaction hash: %v", err)
	}

	pterm.Info.Printfln("transaction hash: %s", txHash)

	return nil
}

func (o *OracleConfig) GetCodeID() (string, error) {
	// Calculate SHA256 hash of the contract file
	contractPath := filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm")
	contractData, err := os.ReadFile(contractPath)
	if err != nil {
		return "", fmt.Errorf("failed to read contract file: %v", err)
	}

	contractHash := fmt.Sprintf("%x", sha256.Sum256(contractData))

	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"q", "wasm", "list-code",
		"--output", "json",
	)

	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to get code id: %v, output: %s", err, output.String())
	}

	var response struct {
		CodeInfos []struct {
			CodeID   string `json:"code_id"`
			Creator  string `json:"creator"`
			DataHash string `json:"data_hash"`
		} `json:"code_infos"`
	}

	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		return "", fmt.Errorf("failed to parse response: %v", err)
	}

	// Look for matching creator and contract hash
	for _, codeInfo := range response.CodeInfos {
		if strings.EqualFold(codeInfo.Creator, o.KeyAddress) &&
			strings.EqualFold(codeInfo.DataHash, contractHash) {
			return codeInfo.CodeID, nil
		}
	}

	return "", nil
}

func (o *OracleConfig) InstantiateContract(rollerData roller.RollappConfig) error {
	instantiateMsg := struct {
		Config struct {
			Updater             string `json:"updater"`
			PriceExpirySeconds  int    `json:"price_expiry_seconds"`
			PriceThresholdRatio string `json:"price_threshold_ratio"`
		} `json:"config"`
	}{
		Config: struct {
			Updater             string `json:"updater"`
			PriceExpirySeconds  int    `json:"price_expiry_seconds"`
			PriceThresholdRatio string `json:"price_threshold_ratio"`
		}{
			Updater:             o.KeyAddress,
			PriceExpirySeconds:  60,
			PriceThresholdRatio: "0.001",
		},
	}

	msgBytes, err := json.Marshal(instantiateMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal instantiate message: %v", err)
	}

	var balanceDenom string
	raResp, err := rollapp.GetMetadataFromChain(rollerData.RollappID, rollerData.HubData)
	if err != nil {
		return fmt.Errorf("failed to get rollapp metadata: %v", err)
	}

	if raResp.Rollapp.GenesisInfo.NativeDenom == nil {
		balanceDenom = consts.Denoms.HubIbcOnRollapp
	} else {
		balanceDenom = raResp.Rollapp.GenesisInfo.NativeDenom.Base
	}

	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"tx", "wasm", "instantiate", o.CodeID,
		string(msgBytes),
		"--from", o.KeyAddress,
		"--label", "price_oracle",
		"--admin", o.KeyAddress,
		"--gas", "auto",
		"--gas-adjustment", "1.3",
		"--fees", fmt.Sprintf("40000000000000000%s", balanceDenom),
		"--keyring-backend", consts.SupportedKeyringBackends.Test.String(),
		"--chain-id", rollerData.RollappID,
		"--broadcast-mode", "sync",
		"--home", o.ConfigDirPath,
		"-y",
	)

	fmt.Println(cmd.String())

	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return fmt.Errorf("failed to instantiate contract: %v, output: %s", err, output)
	}

	time.Sleep(time.Second * 2)

	txHash, err := bash.ExtractTxHash(output.String())
	if err != nil {
		return fmt.Errorf("failed to extract transaction hash: %v", err)
	}

	pterm.Info.Printfln("transaction hash: %s", txHash)

	return nil
}

func (o *OracleConfig) ListContracts(rollerData roller.RollappConfig) ([]string, error) {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"query", "wasm", "list-contracts-by-creator",
		o.KeyAddress,
		"--node", "http://localhost:26657",
		"--chain-id", rollerData.RollappID,
		"--output", "json",
	)

	stdout, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return nil, fmt.Errorf("failed to list contracts: %v", err)
	}

	var response struct {
		Contracts []string `json:"contract_addresses"`
	}

	if err := json.Unmarshal(stdout.Bytes(), &response); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return response.Contracts, nil
}
