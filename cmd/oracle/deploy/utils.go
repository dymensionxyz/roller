package deploy

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/keys"
	"github.com/dymensionxyz/roller/utils/roller"
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

func (o *Oracle) DownloadContractCode() {
	pterm.Error.Println("unimplemented")
}

func (o *Oracle) GetStoreWasmContractCmd() *exec.Cmd {
	return exec.Command(
		consts.Executables.RollappEVM,
		"tx",
		"store",
		filepath.Join(o.ConfigDirPath, "oracle.wasm"),
		"--from",
		o.Key,
		"--chain-id",
		"dymension_100-1",
		"--output",
		"json",
		"--home",
		"/Users/artemijspavlovs/.roller/rollapp/config",
	)
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
	var cmd *exec.Cmd

	switch rollerData.RollappVMType {
	case consts.WASM_ROLLAPP:
		cmd = exec.Command(
			consts.Executables.RollappEVM,
			"tx", "wasm", "store",
			filepath.Join(o.ConfigDirPath, "centralized_oracle.wasm"),
			"--instantiate-anyof-addresses", o.Address,
			"--from", "rol",
			"--gas", "auto",
			"--gas-adjustment", "1.3",
			"--fees", "40693780000000000awasmnat",
		)
	case consts.EVM_ROLLAPP:
		return fmt.Errorf("EVM rollapp type does not support oracle deployment")
	case consts.SDK_ROLLAPP:
		return fmt.Errorf("SDK rollapp type does not support oracle deployment")
	default:
		return fmt.Errorf("unsupported rollapp type: %s", rollerData.RollappVMType)
	}

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to store contract: %v, output: %s", err, output)
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
