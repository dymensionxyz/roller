package rngoracle

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	oracleutils "github.com/dymensionxyz/roller/cmd/oracle/utils"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/roller"
)

//go:embed configs/*
var configFiles embed.FS

type Config struct {
	External struct {
		RandomnessServerBaseURL string `json:"randomness_server_base_url"`
		PollRetryCount          int    `json:"poll_retry_count"`
		PollRetryWaitTime       int64  `json:"poll_retry_wait_time"`
		PollRetryMaxWaitTime    int64  `json:"poll_retry_max_wait_time"`
	} `json:"external"`
	Agent struct {
		HTTPServerAddress string `json:"http_server_address"`
	} `json:"agent"`
	DB struct {
		DBPath string `json:"db_path"`
	} `json:"db"`
	Contract struct {
		PollInterval    int64  `json:"poll_interval"`
		NodeURL         string `json:"node_url"`
		Mnemonic        string `json:"mnemonic"`
		ContractAddress string `json:"contract_address"`
		DerivationPath  string `json:"derivation_path"`
		GasLimit        int64  `json:"gas_limit"`
		GasFeeCap       int64  `json:"gas_fee_cap"`
	} `json:"contract"`
}

func DeployCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploys a price oracle to the RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initconfig.AddFlags(cmd); err != nil {
				pterm.Error.Printf("failed to add flags: %v\n", err)
				return
			}

			if runtime.GOOS != "linux" {
				pterm.Error.Printfln("os %s is not supported", runtime.GOOS)
				return
			}

			home, err := filesystem.ExpandHomePath(
				cmd.Flag(initconfig.GlobalFlagNames.Home).Value.String(),
			)
			if err != nil {
				pterm.Error.Printf("failed to expand home directory: %v\n", err)
				return
			}

			rollerData, err := roller.LoadConfig(home)
			if err != nil {
				pterm.Error.Printf("failed to load roller config file: %v\n", err)
				return
			}

			// Create the appropriate deployer based on RollApp type
			var deployer oracleutils.ContractDeployer
			switch rollerData.RollappVMType {
			case consts.EVM_ROLLAPP:
				deployer, err = oracleutils.NewEVMDeployer(rollerData, consts.Oracles.Price)
				if err != nil {
					pterm.Error.Printf("failed to create evm deployer: %v\n", err)
					return
				}

				err := dependencies.InstallSolidityDependencies()
				if err != nil {
					pterm.Error.Printf("failed to install solidity dependencies: %v\n", err)
					return
				}
			default:
				pterm.Error.Printf("unsupported rollapp type: %s\n", rollerData.RollappVMType)
				return
			}

			contractAddress, isDeployed := deployer.IsContractDeployed()
			if isDeployed {
				pterm.Info.Printf("contract already deployed at: %s\n", contractAddress)
				return
			}

			contracts := []struct {
				Name string
				Url  string
			}{
				{
					Name: "EventManager.sol",
					Url:  "https://storage.googleapis.com/dymension-roller/rng_EventManager.sol",
				},
				{
					Name: "RandomnessGenerator.sol",
					Url:  "https://storage.googleapis.com/dymension-roller/rng_RandomnessGenerator.sol",
				},
			}

			for _, contract := range contracts {
				err = deployer.DownloadContract(contract.Url, contract.Name, consts.Oracles.Rng)
				if err != nil {
					pterm.Error.Printf("failed to download contract: %v\n", err)
					return
				}
			}

			contractAddr, err := deployer.DeployContract(
				context.Background(),
				"RandomnessGenerator.sol",
				consts.Oracles.Rng,
			)
			if err != nil {
				pterm.Error.Printf("failed to deploy contract: %v\n", err)
				return
			}
			pterm.Success.Printf("Contract deployed successfully at: %s\n", contractAddr)

			pterm.Info.Println("starting phase 2: oracle client setup")
			pterm.Info.Println("downloading oracle binary")

			obvi, err := dependencies.GetOracleBinaryVersion(rollerData.RollappVMType)
			if err != nil {
				pterm.Error.Printf("failed to get oracle binary version: %v\n", err)
				return
			}

			var v string
			switch rollerData.RollappVMType {
			case consts.EVM_ROLLAPP:
				v = obvi.RngEvmOracle
			default:
				pterm.Error.Printfln("unsupported rollapp type %s", rollerData.RollappVMType)
				return
			}

			bc := dependencies.BinaryInstallConfig{
				RollappType: rollerData.RollappVMType,
				Version:     v,
				InstallDir:  consts.Executables.RngOracle,
			}

			err = dependencies.InstallBinary(context.Background(), bc, consts.Oracles.Rng)
			if err != nil {
				pterm.Error.Printf("failed to install oracle binary: %v\n", err)
				return
			}

			oracleConfigDir := filepath.Join(
				rollerData.Home,
				consts.ConfigDirName.Oracle,
				consts.Oracles.Rng,
			)
			pterm.Info.Printfln(
				"copying config file into %s",
				oracleConfigDir,
			)
			if err := copyConfigFile(rollerData.RollappVMType, oracleConfigDir); err != nil {
				pterm.Error.Printf("failed to copy config file: %v\n", err)
				return
			}
			pterm.Info.Println("config file copied successfully")
			pterm.Info.Println("updating config values")

			cfp := filepath.Join(oracleConfigDir, "config.json")
			// Read the existing config
			configData, err := os.ReadFile(cfp)
			if err != nil {
				pterm.Error.Printf("failed to read config file: %v\n", err)
				return
			}

			var config Config
			if err := json.Unmarshal(configData, &config); err != nil {
				pterm.Error.Printf("failed to parse config file: %v\n", err)
				return
			}

			// Update the values
			config.Contract.NodeURL = "http://127.0.0.1:8545"
			config.Contract.ContractAddress = contractAddr
			config.Contract.Mnemonic = contractAddr

			// Write back the updated config
			updatedConfig, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				pterm.Error.Printf("failed to marshal updated config: %v\n", err)
				return
			}

			if err := os.WriteFile(cfp, updatedConfig, 0o644); err != nil {
				pterm.Error.Printf("failed to write updated config: %v\n", err)
				return
			}
		},
	}

	return cmd
}

func copyConfigFile(rollappType consts.VMType, destDir string) error {
	var configFile string
	switch rollappType {
	case consts.EVM_ROLLAPP:
		configFile = "configs/evm-config.json"
	default:
		return fmt.Errorf("unsupported rollapp type: %s", rollappType)
	}

	data, err := configFiles.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	configPath := filepath.Join(destDir, "config.json")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
