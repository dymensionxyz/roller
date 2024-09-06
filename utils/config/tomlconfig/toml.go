package tomlconfig

import (
	"encoding/json"
	"os"
	"path/filepath"

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
	var config config.RollappConfig
	tomlBytes, err := os.ReadFile(filepath.Join(root, consts.RollerConfigFileName))
	if err != nil {
		return config, err
	}
	err = naoinatoml.Unmarshal(tomlBytes, &config)
	if err != nil {
		return config, err
	}

	return config, nil
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

func LoadRollappMetadataFromChain(
	home, raID string,
	hd *consts.HubData,
) (*config.RollappConfig, error) {
	var cfg config.RollappConfig
	if hd.ID == "mock" {
		cfg = config.RollappConfig{
			Home:             home,
			RollappID:        raID,
			GenesisHash:      "",
			GenesisUrl:       "",
			RollappBinary:    consts.Executables.RollappEVM,
			VMType:           consts.EVM_ROLLAPP,
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

	if hd.ID != "mock" {
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

		cfg = config.RollappConfig{
			Home:             home,
			GenesisHash:      raResponse.Rollapp.GenesisInfo.GenesisChecksum,
			GenesisUrl:       raResponse.Rollapp.Metadata.GenesisUrl,
			RollappID:        raResponse.Rollapp.RollappId,
			RollappBinary:    consts.Executables.RollappEVM,
			VMType:           consts.EVM_ROLLAPP,
			Denom:            "",
			Decimals:         18,
			HubData:          *hd,
			DA:               consts.DaNetworks["mocha-4"],
			RollerVersion:    "latest",
			Environment:      hd.ID,
			ExecutionVersion: version.BuildVersion,
			Bech32Prefix:     raResponse.Rollapp.GenesisInfo.Bech32Prefix,
			BaseDenom:        "",
			MinGasPrices:     "0",
		}
	}

	return &cfg, nil
}
