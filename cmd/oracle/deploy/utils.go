package deploy

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	cosmossdkmath "cosmossdk.io/math"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

type Oracle struct {
	ConfigDirPath string
	CodeID        string
	ContractAddr  string
	KeyName       string
	KeyAddress    string
}

func NewOracle(rollerData roller.RollappConfig) *Oracle {
	cd := filepath.Join(rollerData.Home, consts.ConfigDirName.Oracle)
	return &Oracle{
		ConfigDirPath: cd,
	}
}

func (o *Oracle) ConfigDir(rollerData roller.RollappConfig) string {
	cd := filepath.Join(rollerData.Home, consts.ConfigDirName.Oracle)
	o.ConfigDirPath = cd

	return o.ConfigDirPath
}

func (o *Oracle) Deploy(rollerData roller.RollappConfig) error {
	pterm.Info.Println("deploying oracle")

	addr, err := generateRaOracleKeys(rollerData.Home, rollerData)
	if err != nil {
		return fmt.Errorf("failed to retrieve oracle keys: %v", err)
	}

	if len(addr) == 0 {
		return fmt.Errorf("no oracle keys generated")
	}

	o.KeyAddress = addr[0].Address
	o.KeyName = addr[0].Name

	j, _ := json.MarshalIndent(o, "", "  ")
	pterm.Info.Println(string(j))
	pterm.Info.Printfln("using oracle key: %s", o.KeyAddress)

	pterm.Info.Println("downloading oracle contract...")
	if err := o.DownloadContractCode(); err != nil {
		return fmt.Errorf("failed to download contract: %v", err)
	}
	pterm.Success.Println("contract downloaded successfully")

	if err := o.GetCodeID(); err != nil {
		if _, ok := err.(*utils.GenericNotFoundError); !ok {
			return fmt.Errorf("failed to get code id: %v", err)
		}

		pterm.Info.Printfln("contract code id not found, creating a new one")
		if err := o.StoreContract(rollerData); err != nil {
			return fmt.Errorf("failed to store contract: %v", err)
		}
	}

	if err := o.GetCodeID(); err != nil {
		return fmt.Errorf("failed to get code id: %v", err)
	}

	pterm.Info.Printfln("contract code id: %s", o.CodeID)

	if err := o.InstantiateContract(rollerData); err != nil {
		return fmt.Errorf("failed to instantiate contract: %v", err)
	}

	pterm.Success.Println("oracle deployed successfully")
	return nil
}

func (o *Oracle) DownloadContractCode() error {
	contractURL := "https://storage.googleapis.com/dymension-roller/centralized_oracle.wasm"
	contractPath := filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm")

	// Create config directory if it doesn't exist
	if err := os.MkdirAll(o.ConfigDirPath, 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %v", err)
	}

	// Create the file
	out, err := os.Create(contractPath)
	if err != nil {
		return fmt.Errorf("failed to create contract file: %v", err)
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(contractURL)
	if err != nil {
		return fmt.Errorf("failed to download contract: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download contract, status: %s", resp.Status)
	}

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save contract: %v", err)
	}

	return nil
}

func generateRaOracleKeys(home string, rollerData roller.RollappConfig) ([]keys.KeyInfo, error) {
	kc := getOracleKeyConfig(rollerData.KeyringBackend)[0]
	ok, err := kc.IsInKeyring(home)
	if err != nil {
		return nil, err
	}

	if ok {
		pterm.Info.Printfln("existing oracle key found, using it")
		ki, err := kc.Info(home)
		if err != nil {
			return nil, err
		}
		return []keys.KeyInfo{*ki}, nil
	}

	shouldImportWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"would you like to import an existing Oracle key?",
	).Show()

	var addr []keys.KeyInfo

	if shouldImportWallet {
		ki, err := kc.Create(home)
		if err != nil {
			return nil, err
		}
		addr = append(addr, *ki)
	} else {
		addr, err = createOraclesKeys(rollerData)
		if err != nil {
			return nil, err
		}
	}

	return addr, nil
}

func createOraclesKeys(rollerData roller.RollappConfig) ([]keys.KeyInfo, error) {
	OracleKeys := getOracleKeyConfig(rollerData.KeyringBackend)
	addresses := make([]keys.KeyInfo, 0)

	for _, key := range OracleKeys {
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

func getOracleKeyConfig(kb consts.SupportedKeyringBackend) []keys.KeyConfig {
	kc := keys.KeyConfig{
		Dir:            consts.ConfigDirName.Oracle,
		ID:             consts.KeysIds.Oracle,
		ChainBinary:    consts.Executables.RollappEVM,
		Type:           consts.SDK_ROLLAPP,
		KeyringBackend: kb,
	}

	return []keys.KeyConfig{kc}
}

func (o *Oracle) StoreContract(rollerData roller.RollappConfig) error {
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
			"--fees", fmt.Sprintf("4000000000000000%s", balanceDenom),
			"--keyring-backend", rollerData.KeyringBackend.String(),
			"--chain-id", rollerData.RollappID,
			"--broadcast-mode", "sync",
			"--home", o.ConfigDirPath,
			"-y",
		)
	case consts.EVM_ROLLAPP:
		return fmt.Errorf("EVM rollapp type does not support oracle deployment")
	case consts.SDK_ROLLAPP:
		return fmt.Errorf("SDK rollapp type does not support oracle deployment")
	default:
		return fmt.Errorf("unsupported rollapp type: %s", rollerData.RollappVMType)
	}

	fmt.Println(cmd.String())
	return errors.New("debug")

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
			kc := getOracleKeyConfig(rollerData.KeyringBackend)[0]
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

func (o *Oracle) GetCodeID() error {
	// Calculate SHA256 hash of the contract file
	contractPath := filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm")
	contractData, err := os.ReadFile(contractPath)
	if err != nil {
		return fmt.Errorf("failed to read contract file: %v", err)
	}

	contractHash := fmt.Sprintf("%x", sha256.Sum256(contractData))

	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"q", "wasm", "list-code",
		"--output", "json",
	)

	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return fmt.Errorf("failed to get code id: %v, output: %s", err, output.String())
	}

	var response struct {
		CodeInfos []struct {
			CodeID   string `json:"code_id"`
			Creator  string `json:"creator"`
			DataHash string `json:"data_hash"`
		} `json:"code_infos"`
	}

	if err := json.Unmarshal(output.Bytes(), &response); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.CodeInfos) == 0 {
		return &utils.GenericNotFoundError{
			Thing: fmt.Sprintf(
				"code with %s as creator and %s as data hash",
				o.KeyAddress,
				contractHash,
			),
		}
	}

	// Look for matching creator and contract hash
	for _, codeInfo := range response.CodeInfos {
		if strings.EqualFold(codeInfo.Creator, o.KeyAddress) &&
			strings.EqualFold(codeInfo.DataHash, contractHash) {
			o.CodeID = codeInfo.CodeID
			return nil
		}
	}

	return nil
}

func (o *Oracle) InstantiateContract(rollerData roller.RollappConfig) error {
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
		"--fees", fmt.Sprintf("4000000000000000%s", balanceDenom),
		"--keyring-backend", rollerData.KeyringBackend.String(),
		"--chain-id", rollerData.RollappID,
		"--broadcast-mode", "sync",
		"--home", o.ConfigDirPath,
		"-y",
	)

	output, err := bash.ExecCommandWithStdout(cmd)
	if err != nil {
		return fmt.Errorf("failed to instantiate contract: %v, output: %s", err, output)
	}

	txHash, err := bash.ExtractTxHash(output.String())
	if err != nil {
		return fmt.Errorf("failed to extract transaction hash: %v", err)
	}

	pterm.Info.Printfln("transaction hash: %s", txHash)

	return nil
}
