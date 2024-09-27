package tomlconfig

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	naoinatoml "github.com/naoina/toml"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/bash"
	"github.com/dymensionxyz/roller/utils/config"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/version"
)

func Write(rlpCfg config.RollappConfig) error {
	tomlBytes, err := naoinatoml.Marshal(rlpCfg)
	if err != nil {
		return err
	}
	// nolint:gofumpt
	return os.WriteFile(filepath.Join(rlpCfg.Home, consts.RollerConfigFileName), tomlBytes, 0o644)
}

// TODO: should be called from root command
func LoadRollerConfig(root string) (config.RollappConfig, error) {
	var rc config.RollappConfig
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
	var config config.RollappConfig
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

func Load(path string) ([]byte, error) {
	tomlBytes, err := os.ReadFile(path)
	if err != nil {
		return tomlBytes, err
	}

	return tomlBytes, nil
}

func GetMockRollappMetadata(
	home, raID string,
	hd *consts.HubData, vmType string,
) (*config.RollappConfig, error) {
	vmt, err := consts.ToVMType(strings.ToLower(vmType))
	if err != nil {
		return nil, err
	}

	cfg := config.RollappConfig{
		Home:             home,
		RollappID:        raID,
		GenesisHash:      "",
		GenesisUrl:       "",
		RollappBinary:    consts.Executables.RollappEVM,
		VMType:           vmt,
		Denom:            "mock",
		Decimals:         18,
		HubData:          *hd,
		DA:               consts.DaNetworks["mock"],
		RollerVersion:    "latest",
		Environment:      "mock",
		ExecutionVersion: version.BuildVersion,
		Bech32Prefix:     "mock",
		BaseDenom:        "amock",
		MinGasPrices:     "0",
	}
	return &cfg, nil
}

func GetRollappMetadataFromChain(
	home, raID string,
	hd *consts.HubData,
) (*config.RollappConfig, error) {
	var cfg config.RollappConfig
	var raResponse rollapp.ShowRollappResponse

	getRollappCmd := rollapp.GetRollappCmd(raID, *hd)

	out, err := bash.ExecCommandWithStdout(getRollappCmd)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(out.Bytes(), &raResponse)
	if err != nil {
		return nil, err
	}

	vmt, _ := consts.ToVMType(strings.ToLower(raResponse.Rollapp.VmType))

	var DA consts.DaData

	switch hd.ID {
	case consts.DevnetHubID:
		DA = consts.DaNetworks[string(consts.CelestiaTestnet)]

	case consts.MainnetHubID:
		DA = consts.DaNetworks[string(consts.CelestiaMainnet)]

	default:
		fmt.Println("unsupported Hub: ", hd.ID)

	}

	cfg = config.RollappConfig{
		Home:             home,
		GenesisHash:      raResponse.Rollapp.GenesisInfo.GenesisChecksum,
		GenesisUrl:       raResponse.Rollapp.Metadata.GenesisUrl,
		RollappID:        raResponse.Rollapp.RollappId,
		RollappBinary:    consts.Executables.RollappEVM,
		VMType:           vmt,
		Denom:            raResponse.Rollapp.GenesisInfo.NativeDenom.Base,
		Decimals:         18,
		HubData:          *hd,
		DA:               DA,
		RollerVersion:    "latest",
		Environment:      hd.ID,
		ExecutionVersion: version.BuildVersion,
		Bech32Prefix:     raResponse.Rollapp.GenesisInfo.Bech32Prefix,
		BaseDenom:        "",
		MinGasPrices:     "0",
	}

	return &cfg, nil
}
