package roller

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
)

// RollappConfig struct represents the information for creating roller.toml  config file
func PrintTokenSupplyLine(rollappConfig RollappConfig) {
	pterm.DefaultSection.WithIndentCharacter("💰").Printf(
		"Total Token Supply: %s %s.",
		addCommasToNum(
			consts.DefaultTokenSupply[:len(consts.DefaultTokenSupply)-int(rollappConfig.Decimals)],
		),
		rollappConfig.Denom,
	)

	pterm.DefaultBasicText.Printf(
		"Note that 1 %s == 1 * 10^%d %s (like 1 ETH == 1 * 10^18 wei).\nThe total supply in base denom (%s) is "+pterm.Yellow(
			"%s%s",
		),
		rollappConfig.Denom,
		rollappConfig.Decimals,
		rollappConfig.BaseDenom,
		rollappConfig.BaseDenom,
		consts.DefaultTokenSupply,
		rollappConfig.BaseDenom,
	)
}

func addCommasToNum(number string) string {
	var result strings.Builder
	n := len(number)

	for i := 0; i < n; i++ {
		if i > 0 && (n-i)%3 == 0 {
			result.WriteString(",")
		}
		result.WriteByte(number[i])
	}
	return result.String()
}

func WriteConfigToDisk(
	c RollappConfig,
) error {
	rollerConfigFilePath := filepath.Join(c.Home, consts.RollerConfigFileName)
	err := tomlconfig.WriteConfigToTOML(c, rollerConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to dump config to TOML: %w", err)
	}

	return nil
}

func (c RollappConfig) ValidateConfig() error {
	err := VerifyHubData(c.HubData)
	if err != nil {
		return err
	}

	if c.BaseDenom == "" {
		return fmt.Errorf("base denom should be populated")
	}

	if !IsValidDAType(string(c.DA.Backend)) {
		return fmt.Errorf("invalid DA type: %s. supported types %s", c.DA.Backend, SupportedDas)
	}

	return nil
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

	settlementNodeAddress, err := tomlconfig.GetKeyFromFile(
		dymintConfigPath,
		"settlement_node_address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current settlement node address", err)
		return nil, err
	}

	rollappMinimumGasPrice, err := tomlconfig.GetKeyFromFile(
		appConfigPath,
		"minimum-gas-prices",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current minimum gas price", err)
		return nil, err
	}

	apiAddress, err := tomlconfig.GetKeyFromFile(
		appConfigPath,
		"api.address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current rest api addr", err)
		return nil, err
	}

	jsonRpcAddress, err := tomlconfig.GetKeyFromFile(
		appConfigPath,
		"json-rpc.address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current settlement json-rpc addr", err)
		return nil, err
	}

	wsAddress, err := tomlconfig.GetKeyFromFile(
		appConfigPath,
		"json-rpc.ws-address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current json-rpc addr ", err)
		return nil, err
	}

	grpcAddress, err := tomlconfig.GetKeyFromFile(
		appConfigPath,
		"grpc-web.address",
	)
	if err != nil {
		pterm.Error.Println("failed to get the current grpc-web addr", err)
		return nil, err
	}

	rpcAddr, err := tomlconfig.GetKeyFromFile(
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

// FindHubDataByID is intended to retrieve consts.HubData from consts.Hubs @20240927
func FindHubDataByID(
	hubs map[string]consts.HubData,
	chainID string,
) (string, consts.HubData, bool) {
	for key, hubData := range hubs {
		if hubData.ID == chainID {
			return key, hubData, true
		}
	}
	return "", consts.HubData{}, false
}
