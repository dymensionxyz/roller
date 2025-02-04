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
				deployer, err = oracleutils.NewEVMDeployer(rollerData, consts.Oracles.Rng)
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

			obvi, err := dependencies.GetOracleBinaryVersion(rollerData.RollappVMType)
			if err != nil {
				pterm.Error.Printf("failed to get oracle binary version: %v", err)
				return
			}
			pterm.Info.Println("downloading oracle binaries")
			err = initRngOracleClient(rollerData, contractAddr, deployer.Mnemonic(), obvi)
			if err != nil {
				pterm.Error.Printf("failed to initialize oracle client: %v\n", err)
				return
			}

			pterm.Info.Println("starting phase 3: rng service setup")
			bc := dependencies.BinaryInstallConfig{
				RollappType: rollerData.RollappVMType,
				Version:     obvi.RngEvmRandomService,
				InstallDir:  consts.Executables.RngOracleRandomService,
			}
			if err := dependencies.InstallRngServiceBinary(context.Background(), bc, consts.Oracles.Rng); err != nil {
				pterm.Error.Printf("failed to initialize rng service client: %v\n", err)
				return
			}

		},
	}

	return cmd
}

func initRngOracleClient(
	rollerData roller.RollappConfig,
	contractAddr, mnemonic string,
	obvi *dependencies.OracleBinaryVersionInfo,
) error {
	var v string
	switch rollerData.RollappVMType {
	case consts.EVM_ROLLAPP:
		v = obvi.RngEvmOracle
	default:
		return fmt.Errorf("unsupported rollapp type %s", rollerData.RollappVMType)
	}

	bc := dependencies.BinaryInstallConfig{
		RollappType: rollerData.RollappVMType,
		Version:     v,
		InstallDir:  consts.Executables.RngOracle,
	}

	if err := dependencies.InstallOracleBinary(context.Background(), bc, consts.Oracles.Rng); err != nil {
		return fmt.Errorf("failed to install oracle binary: %v", err)
	}

	oracleConfigDir := filepath.Join(
		rollerData.Home,
		consts.ConfigDirName.Oracle,
		consts.Oracles.Rng,
	)
	pterm.Info.Printfln("copying config file into %s", oracleConfigDir)
	if err := copyConfigFile(rollerData.RollappVMType, oracleConfigDir); err != nil {
		return fmt.Errorf("failed to copy config file: %v", err)
	}
	pterm.Info.Println("config file copied successfully")
	pterm.Info.Println("updating config values")

	cfp := filepath.Join(oracleConfigDir, "config.json")
	configData, err := os.ReadFile(cfp)
	if err != nil {
		return fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := json.Unmarshal(configData, &config); err != nil {
		return fmt.Errorf("failed to parse config file: %v", err)
	}

	config.Contract.NodeURL = "http://127.0.0.1:8545"
	config.Contract.ContractAddress = contractAddr
	config.Contract.Mnemonic = mnemonic
	config.DB.DBPath = filepath.Join(oracleConfigDir, "db")

	updatedConfig, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %v", err)
	}

	if err := os.WriteFile(cfp, updatedConfig, 0o644); err != nil {
		return fmt.Errorf("failed to write updated config: %v", err)
	}

	return nil
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
