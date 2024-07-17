package initrollapp

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/dymensionxyz/roller/cmd/utils"
)

type RollerConfig struct {
	BaseDenom        string `toml:"base_denom"`
	Bech32Prefix     string `toml:"bech32_prefix"`
	Da               string `toml:"da"`
	Denom            string `toml:"denom"`
	DenomExponent    string `toml:"denom_exponent"`
	DenomLogoDataURI string `toml:"denom_logo_data_uri"`
	Environment      string `toml:"environment"`
	Execution        string `toml:"execution"`
	ExecutionVersion string `toml:"execution_version"`
	Home             string `toml:"home"`
	LogoDataURI      string `toml:"logo_data_uri"`
	MinimumGasPrices string `toml:"minimum_gas_prices"`
	RollappID        string `toml:"rollapp_id"`
	RollerVersion    string `toml:"roller_version"`
}

func NewMockRollerConfig() *RollerConfig {
	return &RollerConfig{
		BaseDenom:        "amock",
		Bech32Prefix:     "ethm",
		Da:               "local",
		Denom:            "mock",
		DenomExponent:    "18",
		DenomLogoDataURI: "",
		Environment:      "local",
		Execution:        "evm",
		ExecutionVersion: "v2.2.0-hotfix.1",
		Home:             "",
		LogoDataURI:      "",
		MinimumGasPrices: "1000000000amock",
		RollappID:        "mockrollapp_100000-1",
		RollerVersion:    "1.0.5",
	}
}

func WriteMockRollerconfigToFile(rc *RollerConfig) error {
	home := utils.GetRollerRootDir()
	filePath := filepath.Join(home, "roller.toml")

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}

	// nolint S1023
	defer file.Close()

	if err := toml.NewEncoder(file).Encode(rc); err != nil {
		fmt.Println("failed to encode toml file", err)
	}

	return nil
}
