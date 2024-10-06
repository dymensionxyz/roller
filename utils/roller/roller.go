package roller

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	naoinatoml "github.com/naoina/toml"
	"github.com/pterm/pterm"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/version"
)

func GetRootDir() string {
	return filepath.Join(os.Getenv("HOME"), ".roller")
}

func GetConfigPath(home string) string {
	return filepath.Join(home, "roller.toml")
}

func CreateConfigFile(home string) (bool, error) {
	rollerConfigFilePath := GetConfigPath(home)

	_, err := os.Stat(rollerConfigFilePath)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			pterm.Info.Println("roller.toml not found, creating")
			_, err := os.Create(rollerConfigFilePath)
			if err != nil {
				pterm.Error.Printf(
					"failed to create %s: %v", rollerConfigFilePath, err,
				)
				return false, err
			}

			return true, nil
		}
	}

	return false, nil
}

// TODO: should be called from root command
func LoadRollerConfig(root string) (RollappConfig, error) {
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
