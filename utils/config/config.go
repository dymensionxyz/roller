package config

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/dymensionxyz/roller/cmd/consts"
	globalutils "github.com/dymensionxyz/roller/utils"
	"github.com/pterm/pterm"
)

var SupportedDas = []consts.DAType{consts.Celestia, consts.Avail, consts.Local}

type RollappConfig struct {
	Home          string        `toml:"home"`
	GenesisHash   string        `toml:"genesis_hash"`
	GenesisUrl    string        `toml:"genesis_url"`
	RollappID     string        `toml:"rollapp_id"`
	RollappBinary string        `toml:"rollapp_binary"`
	VMType        consts.VMType `toml:"execution"`
	Denom         string        `toml:"denom"`
	// TokenSupply   string
	Decimals      uint
	HubData       consts.HubData
	DA            consts.DaData
	RollerVersion string `toml:"roller_version"`

	// new roller.toml
	Environment string `toml:"environment"`
	// Execution        string `toml:"execution"`
	ExecutionVersion string `toml:"execution_version"`
	Bech32Prefix     string `toml:"bech32_prefix"`
	BaseDenom        string `toml:"base_denom"`
	MinGasPrices     string `toml:"minimum_gas_prices"`
}

func (c RollappConfig) Validate() error {
	err := VerifyHubData(c.HubData)
	if err != nil {
		return err
	}

	// the assumption is that the supply is coming from the genesis creator
	// err = VerifyTokenSupply(c.TokenSupply)
	// if err != nil {
	// 	return err
	// }

	if !IsValidDAType(string(c.DA.Backend)) {
		fmt.Println(c.DA.Backend)
		return fmt.Errorf("invalid DA type: %s. supported types %s", c.DA, SupportedDas)
	}

	return nil
}

func IsValidDAType(t string) bool {
	switch consts.DAType(t) {
	case consts.Local, consts.Celestia, consts.Avail:
		return true
	}
	return false
}

func IsValidVMType(t string) bool {
	switch consts.VMType(t) {
	case consts.SDK_ROLLAPP, consts.EVM_ROLLAPP:
		return true
	}
	return false
}

func VerifyHubData(data consts.HubData) error {
	if data.ID == "mock" {
		return nil
	}

	if data.ID == "" {
		return fmt.Errorf("invalid hub id: %s. ID cannot be empty", data.ID)
	}

	if data.RPC_URL == "" {
		return fmt.Errorf("invalid RPC endpoint: %s. RPC URL cannot be empty", data.ID)
	}
	return nil
}

// func VerifyTokenSupply(supply string) error {
// 	tokenSupply := new(big.Int)
// 	_, ok := tokenSupply.SetString(supply, 10)
// 	if !ok {
// 		return fmt.Errorf("invalid token supply: %s. Must be a valid integer", supply)
// 	}
//
// 	ten := big.NewInt(10)
// 	remainder := new(big.Int)
// 	remainder.Mod(tokenSupply, ten)
//
// 	if remainder.Cmp(big.NewInt(0)) != 0 {
// 		return fmt.Errorf("invalid token supply: %s. Must be divisible by 10", supply)
// 	}
//
// 	if tokenSupply.Cmp(big.NewInt(10_000_000)) < 0 {
// 		return fmt.Errorf("token supply %s must be greater than 10,000,000", tokenSupply)
// 	}
//
// 	return nil
// }

func ValidateDecimals(decimals uint) error {
	if decimals > 18 {
		return fmt.Errorf("invalid decimals: %d. Must be less than or equal to 18", decimals)
	}
	return nil
}

func IsValidDenom(s string) error {
	if !strings.HasPrefix(s, "a") {
		return fmt.Errorf("invalid denom '%s'. denom expected to start with 'a'", s)
	}
	if !IsValidTokenSymbol(s[1:]) {
		return fmt.Errorf("invalid token symbol '%s'", s[1:])
	}
	return nil
}

func IsValidTokenSymbol(s string) bool {
	if len(s) < 3 || len(s) > 6 {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) ||
			!strings.ContainsRune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", r) {
			return false
		}
	}
	return true
}

func GetConfigurableRollappValues(home string) (map[string]string, error) {
	dymintConfigPath := filepath.Join(
		home,
		consts.ConfigDirName.Rollapp,
		"config",
		"dymint.toml",
	)
	appConfigPath := filepath.Join(
		home,
		consts.ConfigDirName.Rollapp,
		"config",
		"app.toml",
	)
	// nice name, ik
	configConfigPath := filepath.Join(
		home,
		consts.ConfigDirName.Rollapp,
		"config",
		"config.toml",
	)

	settlementNodeAddress, err := globalutils.GetKeyFromTomlFile(
		dymintConfigPath,
		"settlement_node_address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current settlement node address", err)
		return nil, err
	}

	rollappMinimumGasPrice, err := globalutils.GetKeyFromTomlFile(
		appConfigPath,
		"minimum-gas-prices",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current minimum gas price", err)
		return nil, err
	}

	apiAddress, err := globalutils.GetKeyFromTomlFile(
		appConfigPath,
		"api.address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current rest api addr", err)
		return nil, err
	}

	jsonRpcAddress, err := globalutils.GetKeyFromTomlFile(
		appConfigPath,
		"json-rpc.address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current settlement json-rpc addr", err)
		return nil, err
	}

	wsAddress, err := globalutils.GetKeyFromTomlFile(
		appConfigPath,
		"json-rpc.ws-address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current json-rpc addr ", err)
		return nil, err
	}

	grpcAddress, err := globalutils.GetKeyFromTomlFile(
		appConfigPath,
		"grpc-web.address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current grpc-web addr", err)
		return nil, err
	}

	rpcAddr, err := globalutils.GetKeyFromTomlFile(
		configConfigPath,
		"rpc.laddr",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current rpc addr", err)
		return nil, err
	}

	values := map[string]string{
		"rollapp_minimum_gas_price": rollappMinimumGasPrice,
		"rollapp_rpc_port":          rpcAddr,
		"rollapp_grpc_port":         grpcAddress,
		"rollapp_rest_api_port":     apiAddress,
		"rollapp_json_rpc_port":     jsonRpcAddress,
		"rollapp_ws_port":           wsAddress,
		"settlement_node_address":   settlementNodeAddress,
		"da_node_address":           "",
	}

	return values, nil
}

func TableDataFromMap(values map[string]string) ([][]string, error) {
	tableData := [][]string{
		{"Key", "Current Value"}, // Header row
	}

	for k, v := range values {
		tableData = append(tableData, []string{k, v})
	}

	return tableData, nil
}

func ShowCurrentConfigurableValues(home string) error {
	cv, err := GetConfigurableRollappValues(home)
	if err != nil {
		return err
	}
	td, err := TableDataFromMap(cv)
	if err != nil {
		return err
	}
	err = pterm.DefaultTable.WithHasHeader().WithData(td).Render()
	if err != nil {
		fmt.Printf("Error rendering table: %v\n", err)
		return err
	}

	return nil
}

type PathValue struct {
	Path  string
	Value interface{}
}
