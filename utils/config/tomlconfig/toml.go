package tomlconfig

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"

	comettypes "github.com/cometbft/cometbft/types"
	naoinatoml "github.com/naoina/toml"
	"github.com/pterm/pterm"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
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
			RollappID:        "mock_1000-1",
			GenesisHash:      "",
			GenesisUrl:       "",
			RollappBinary:    consts.Executables.RollappEVM,
			VMType:           consts.EVM_ROLLAPP,
			Denom:            "mock",
			Decimals:         18,
			HubData:          *hd,
			DA:               "local",
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
		getRollappCmd := exec.Command(
			consts.Executables.Dymension,
			"q", "rollapp", "show",
			raID, "-o", "json", "--node", hd.RPC_URL,
		)

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
			GenesisHash:      raResponse.Rollapp.GenesisChecksum,
			GenesisUrl:       raResponse.Rollapp.Metadata.GenesisUrl,
			RollappID:        raResponse.Rollapp.RollappId,
			RollappBinary:    consts.Executables.RollappEVM,
			VMType:           consts.EVM_ROLLAPP,
			Denom:            "mock",
			Decimals:         18,
			HubData:          *hd,
			DA:               consts.Celestia,
			RollerVersion:    "latest",
			Environment:      hd.ID,
			ExecutionVersion: version.BuildVersion,
			Bech32Prefix:     raResponse.Rollapp.Bech32Prefix,
			BaseDenom:        "amock",
			MinGasPrices:     "0",
		}

		genesisPath := initconfig.GetGenesisFilePath(home)
		genesisUrl := raResponse.Rollapp.Metadata.GenesisUrl
		err = downloadFile(genesisUrl, genesisPath)
		if err != nil {
			return nil, err
		}

		// move to helper function with a spinner?
		genesis, err := comettypes.GenesisDocFromFile(genesisPath)
		if err != nil {
			return nil, err
		}

		if genesis.ChainID != raID {
			err = fmt.Errorf(
				"the genesis file ChainID (%s) does not match  the rollapp ID you're trying to initialize ("+
					"%s)",
				genesis.ChainID,
				raID,
			)
			return nil, err
		}

		downloadedGenesisHash, err := calculateSHA256(genesisPath)
		if err != nil {
			pterm.Error.Println("failed to calculate hash of genesis file: ", err)
		}
		raGenesisHash, _ := getRollappGenesisHash(raID, *hd)
		if downloadedGenesisHash != raGenesisHash {
			err = errors.New(
				"the hash of the downloaded file does not match the one registered with the rollapp",
			)
			return nil, err
		}
	}

	return &cfg, nil
}

// TODO: download the file in chunks if possible
func downloadFile(url, filepath string) error {
	spinner, _ := pterm.DefaultSpinner.
		Start("Downloading genesis file from ", url)

	// nolint:gosec
	resp, err := http.Get(url)
	if err != nil {
		spinner.Fail("failed to download file: ", err)
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(filepath)
	if err != nil {
		spinner.Fail("failed to download file: ", err)
		return err
	}
	defer out.Close()

	spinner.Success("Successfully downloaded the genesis file")
	_, err = io.Copy(out, resp.Body)
	return err
}

func calculateSHA256(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		return "", fmt.Errorf("error opening file: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("error calculating hash: %v", err)
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func getRollappGenesisHash(raID string, hd consts.HubData) (string, error) {
	var raResponse rollapp.ShowRollappResponse
	getRollappCmd := exec.Command(
		consts.Executables.Dymension,
		"q", "rollapp", "show",
		raID, "-o", "json", "--node", hd.RPC_URL,
	)

	out, err := bash.ExecCommandWithStdout(getRollappCmd)
	if err != nil {
		return "", err
	}

	err = json.Unmarshal(out.Bytes(), &raResponse)
	if err != nil {
		return "", err
	}
	return raResponse.Rollapp.GenesisChecksum, nil
}
