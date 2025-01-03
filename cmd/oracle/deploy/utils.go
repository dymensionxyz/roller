package deploy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
	"github.com/dymensionxyz/roller/utils/tx"
)

type Oracle struct {
	ConfigDirPath string
	Key           string
	CodeID        string
	ContractAddr  string
	Address       string
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

	addr, err := generateRaOracleKeys(o.ConfigDirPath, rollerData)
	if err != nil {
		return fmt.Errorf("failed to retrieve oracle keys: %v", err)
	}

	if len(addr) == 0 {
		return fmt.Errorf("no oracle keys generated")
	}

	o.Address = addr[0].Address
	pterm.Info.Printfln("using oracle key: %s", o.Address)

	pterm.Info.Println("downloading oracle contract...")
	if err := o.DownloadContractCode(); err != nil {
		return fmt.Errorf("failed to download contract: %v", err)
	}
	pterm.Success.Println("contract downloaded successfully")

	if err := o.StoreContract(rollerData); err != nil {
		return fmt.Errorf("failed to store contract: %v", err)
	}

	if err := o.GetCodeID(); err != nil {
		return fmt.Errorf("failed to get code id: %v", err)
	}

	pterm.Info.Printfln("contract code id: %s", o.CodeID)

	if err := o.InstantiateContract(); err != nil {
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
	useExistingOracleWallet, _ := pterm.DefaultInteractiveConfirm.WithDefaultText(
		"would you like to import an existing Oracle key?",
	).Show()

	var addr []keys.KeyInfo
	var err error

	if useExistingOracleWallet {
		kc, err := keys.NewKeyConfig(
			consts.ConfigDirName.Oracle,
			consts.KeysIds.Oracle,
			consts.Executables.RollappEVM,
			consts.SDK_ROLLAPP,
			rollerData.KeyringBackend,
			keys.WithRecover(),
		)
		if err != nil {
			return nil, err
		}

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
	}
	return addresses, nil
}

func getOracleKeyConfig(kb consts.SupportedKeyringBackend) []keys.KeyConfig {
	kc := keys.KeyConfig{
		Dir:            consts.ConfigDirName.HubKeys,
		ID:             consts.KeysIds.Oracle,
		ChainBinary:    consts.Executables.Dymension,
		Type:           consts.SDK_ROLLAPP,
		KeyringBackend: kb,
	}

	return []keys.KeyConfig{kc}
}

func (o *Oracle) StoreContract(rollerData roller.RollappConfig) error {
	var cmdName string
	var args []string

	switch rollerData.RollappVMType {
	case consts.WASM_ROLLAPP:
		cmdName = consts.Executables.RollappEVM
		args = []string{
			"tx", "wasm", "store",
			filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm"),
			"--instantiate-anyof-addresses", o.Address,
			"--from", "rol",
			"--gas", "auto",
			"--gas-adjustment", "1.3",
			"--fees", "40693780000000000awasmnat",
		}
	case consts.EVM_ROLLAPP:
		return fmt.Errorf("EVM rollapp type does not support oracle deployment")
	case consts.SDK_ROLLAPP:
		return fmt.Errorf("SDK rollapp type does not support oracle deployment")
	default:
		return fmt.Errorf("unsupported rollapp type: %s", rollerData.RollappVMType)
	}

	// Prepare prompt responses for automatic handling
	var promptResponses map[string]string
	if rollerData.KeyringBackend == consts.SupportedKeyringBackends.OS {
		pswFileName, err := filesystem.GetOsKeyringPswFileName(consts.Executables.RollappEVM)
		if err != nil {
			return err
		}
		fp := filepath.Join(rollerData.Home, string(pswFileName))
		psw, err := filesystem.ReadFromFile(fp)
		if err != nil {
			return err
		}

		promptResponses = map[string]string{
			"Enter keyring passphrase":    psw,
			"Re-enter keyring passphrase": psw,
		}
	} else {
		promptResponses = map[string]string{}
	}

	manualPrompts := map[string]string{
		"signatures": fmt.Sprintf(
			"this transaction will store the oracle contract byte code on chain with %s as the updater. do you want to continue?",
			o.Address,
		),
	}

	output, err := bash.ExecuteCommandWithPromptHandler(
		cmdName,
		args,
		promptResponses,
		manualPrompts,
	)
	if err != nil {
		return fmt.Errorf("failed to store contract: %v, output: %s", err, output)
	}

	// Extract transaction hash
	txHash, err := bash.ExtractTxHash(output.String())
	if err != nil {
		return fmt.Errorf("failed to extract transaction hash: %v", err)
	}

	pterm.Info.Printfln("transaction hash: %s", txHash)

	// Monitor transaction
	wsURL := "ws://localhost:26657"
	if err := tx.MonitorTransaction(wsURL, txHash); err != nil {
		return fmt.Errorf("failed to monitor transaction: %v", err)
	}

	return nil
}

func (o *Oracle) GetCodeID() error {
	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"q", "wasm", "list-code",
		"--output", "json",
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to get code id: %v", err)
	}

	var response struct {
		CodeInfos []struct {
			CodeID string `json:"code_id"`
		} `json:"code_infos"`
	}

	if err := json.Unmarshal(output, &response); err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	if len(response.CodeInfos) == 0 {
		return fmt.Errorf("no code found")
	}

	o.CodeID = response.CodeInfos[0].CodeID
	return nil
}

func (o *Oracle) InstantiateContract() error {
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
			Updater:             o.Address,
			PriceExpirySeconds:  60,
			PriceThresholdRatio: "0.001",
		},
	}

	msgBytes, err := json.Marshal(instantiateMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal instantiate message: %v", err)
	}

	cmd := exec.Command(
		consts.Executables.RollappEVM,
		"tx", "wasm", "instantiate", o.CodeID,
		string(msgBytes),
		"--from", o.Address,
		"--label", "price_oracle",
		"--gas", "auto",
		"--gas-adjustment", "1.3",
		"--fees", "40693780000000000awasmnat",
		"--admin", o.Address,
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to instantiate contract: %v, output: %s", err, output)
	}

	return nil
}