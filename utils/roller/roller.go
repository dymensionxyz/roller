package roller

import (
	"os"
	"path/filepath"
	"strings"

	naoinatoml "github.com/naoina/toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/version"
)

// GetRootDir returns the root directory for roller configuration (~/.roller)
func GetRootDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".roller")
}

func GetConfigPath(home string) string {
	return filepath.Join(home, consts.RollerConfigFileName)
}

func CreateConfigFileIfNotPresent(home string) (bool, error) {
	rollerConfigFilePath := GetConfigPath(home)
	ok, err := filesystem.DoesFileExist(rollerConfigFilePath)
	if err != nil {
		pterm.Error.Println("failed to check roller config file existence", err)
		return false, err
	}

	if !ok {
		pterm.Info.Printf("%s does not exist, creating...\n", rollerConfigFilePath)
		_, err := os.Create(rollerConfigFilePath)
		if err != nil {
			pterm.Error.Printf(
				"failed to create %s: %v", rollerConfigFilePath, err,
			)
			return false, err
		}

		return true, nil
	}

	return true, nil
}

// TODO: should be called from root command
func LoadConfig(root string) (RollappConfig, error) {
	var rc RollappConfig
	tomlBytes, err := os.ReadFile(filepath.Join(root, consts.RollerConfigFileName))
	if err != nil {
		return rc, err
	}
	err = naoinatoml.Unmarshal(tomlBytes, &rc)
	if err != nil {
		return rc, err
	}

	return rc, nil
}

func WriteConfig(rlpCfg RollappConfig) error {
	tomlBytes, err := naoinatoml.Marshal(rlpCfg)
	if err != nil {
		return err
	}
	// nolint:gofumpt
	return os.WriteFile(filepath.Join(rlpCfg.Home, consts.RollerConfigFileName), tomlBytes, 0o644)
}

func LoadHubData(root string) (consts.HubData, error) {
	var config RollappConfig
	tomlBytes, err := os.ReadFile(filepath.Join(root, consts.RollerConfigFileName))
	if err != nil {
		return config.HubData, err
	}
	err = naoinatoml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config.HubData, err
	}

	return config.HubData, nil
}

func GetMockRollappMetadata(
	home, raID string,
	hd *consts.HubData, vmType string,
) (*RollappConfig, error) {
	vmt, err := consts.ToVMType(strings.ToLower(vmType))
	if err != nil {
		return nil, err
	}

	cfg := RollappConfig{
		Home:                 home,
		RollappID:            raID,
		KeyringBackend:       "test",
		GenesisHash:          "",
		GenesisUrl:           "",
		RollappBinary:        consts.Executables.RollappEVM,
		RollappVMType:        vmt,
		Denom:                "mock",
		Decimals:             18,
		HubData:              *hd,
		DA:                   consts.DaNetworks["mock"],
		RollerVersion:        "latest",
		Environment:          "mock",
		RollappBinaryVersion: version.BuildVersion,
		Bech32Prefix:         "mock",
		BaseDenom:            "amock",
		MinGasPrices:         "0",
	}
	return &cfg, nil
}
