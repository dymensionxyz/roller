package initrollapp

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/utils/filesystem"
)

type RollerConfig struct {
	BaseDenom            string `toml:"base_denom"`
	Bech32Prefix         string `toml:"bech32_prefix"`
	Da                   string `toml:"da"`
	Denom                string `toml:"denom"`
	DenomExponent        string `toml:"denom_exponent"`
	DenomLogoDataURI     string `toml:"denom_logo_data_uri"`
	Environment          string `toml:"environment"`
	RollappVMType        string `toml:"rollapp_vm_type"`
	RollappBinaryVersion string `toml:"rollapp_binary_version"`
	Home                 string `toml:"home"`
	LogoDataURI          string `toml:"logo_data_uri"`
	MinimumGasPrices     string `toml:"minimum_gas_prices"`
	RollappID            string `toml:"rollapp_id"`
	RollerVersion        string `toml:"roller_version"`
}

func NewMockRollerConfig(cmd *cobra.Command) *RollerConfig {
	home, _ := filesystem.ExpandHomePath(cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String())

	return &RollerConfig{
		BaseDenom:            "amock",
		Bech32Prefix:         "ethm",
		Da:                   "local",
		Denom:                "mock",
		DenomExponent:        "18",
		DenomLogoDataURI:     "",
		Environment:          "local",
		RollappVMType:        "evm",
		RollappBinaryVersion: "v2.2.0-hotfix.1",
		Home:                 home,
		LogoDataURI:          "",
		MinimumGasPrices:     "1000000000amock",
		RollappID:            "mockrollapp_100000-1",
		RollerVersion:        "1.0.5",
	}
}

func WriteMockRollerconfigToFile(rc *RollerConfig) error {
	home := rc.Home
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
