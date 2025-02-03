package priceoracle

import (
	"context"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	cosmossdkmath "cosmossdk.io/math"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	oracleutils "github.com/dymensionxyz/roller/cmd/oracle/utils"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

//go:embed configs/*
var configFiles embed.FS

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
			var contractUrl string
			switch rollerData.RollappVMType {
			case consts.EVM_ROLLAPP:
				deployer, err = oracleutils.NewEVMDeployer(rollerData, consts.Oracles.Price)
				if err != nil {
					pterm.Error.Printf("failed to create evm deployer: %v\n", err)
					return
				}
				contractUrl = "https://storage.googleapis.com/dymension-roller/price_oracle_contract.sol"

				err := dependencies.InstallSolidityDependencies()
				if err != nil {
					pterm.Error.Printf("failed to install solidity dependencies: %v\n", err)
					return
				}

				err = deployer.DownloadContract(
					contractUrl,
					"PriceOracle.sol",
					consts.Oracles.Price,
				)
				if err != nil {
					pterm.Error.Printf("failed to download contract: %v\n", err)
					return
				}
			case consts.WASM_ROLLAPP:
				deployer, err = oracleutils.NewWasmDeployer(rollerData, consts.Oracles.Price)
				if err != nil {
					pterm.Error.Printf("failed to create wasm deployer: %v\n", err)
					return
				}
				contractUrl = "https://storage.googleapis.com/dymension-roller/price_oracle_contract.wasm"

				err = deployer.DownloadContract(
					contractUrl,
					"PriceOracle.wasm",
					consts.Oracles.Price,
				)
				if err != nil {
					pterm.Error.Printf("failed to download contract: %v\n", err)
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

			if err != nil {
				pterm.Error.Printf("failed to create deployer: %v\n", err)
				return
			}

			contractAddr, err := deployer.DeployContract(
				context.Background(),
				"PriceOracle.sol",
				consts.Oracles.Price,
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
				v = obvi.PriceEvmOracle
			case consts.WASM_ROLLAPP:
				v = obvi.PriceWasmOracle
			default:
				pterm.Error.Printfln("unsupported rollapp type %s", rollerData.RollappVMType)
				return
			}

			bc := dependencies.BinaryInstallConfig{
				RollappType: rollerData.RollappVMType,
				Version:     v,
				InstallDir:  consts.Executables.PriceOracle,
			}

			err = dependencies.InstallBinary(context.Background(), bc, consts.Oracles.Price)
			if err != nil {
				pterm.Error.Printf("failed to install oracle binary: %v\n", err)
				return
			}

			oracleConfigDir := filepath.Join(
				rollerData.Home,
				consts.ConfigDirName.Oracle,
				consts.Oracles.Price,
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
			gl, _ := cosmossdkmath.NewIntFromString(
				consts.DefaultMinGasPrice,
			)

			raData, err := rollapp.GetMetadataFromChain(rollerData.RollappID, rollerData.HubData)
			if err != nil {
				pterm.Error.Printf("failed to get rollapp metadata: %v\n", err)
				return
			}

			var feeDenom string
			if raData.Rollapp.GenesisInfo.NativeDenom == nil {
				feeDenom = consts.Denoms.HubIbcOnRollapp
			} else {
				feeDenom = raData.Rollapp.GenesisInfo.NativeDenom.Base
			}

			var updates map[string]any

			switch rollerData.RollappVMType {
			case consts.EVM_ROLLAPP:
				networkID, err := oracleutils.ExtractNetworkID(rollerData.RollappID)
				if err != nil {
					pterm.Error.Printf("failed to extract network ID: %v\n", err)
					return
				}

				updates = map[string]any{
					"chainClient.rpcEndpoint":     "http://127.0.0.1:8545/",
					"chainClient.chainId":         networkID,
					"chainClient.privateKey":      deployer.PrivateKey(),
					"chainClient.contractAddress": contractAddr,
					"grpc_port":                   9093,
					// gasLimit: 250000
					// maxGasPrice: "100000000000"
				}
			case consts.WASM_ROLLAPP:
				updates = map[string]any{
					"chainClient.oracleContractAddress": contractAddr,
					"chainClient.fee": fmt.Sprintf(
						"%s%s",
						"40000000000000000000",
						feeDenom,
					),
					"chainClient.gasLimit":      gl.Uint64(),
					"chainClient.bech32Prefix":  raData.Rollapp.GenesisInfo.Bech32Prefix,
					"chainClient.chainId":       raData.Rollapp.RollappId,
					"chainClient.privateKey":    deployer.PrivateKey(),
					"chainClient.ssl":           false,
					"chainClient.chainGrpcHost": "localhost:9090",
					"grpc_port":                 9093,
				}
			default:
				pterm.Error.Printf("unsupported rollapp type: %s\n", rollerData.RollappVMType)
				return
			}

			cfp := filepath.Join(oracleConfigDir, "config.yaml")
			err = yamlconfig.UpdateNestedYAML(cfp, updates)
			if err != nil {
				pterm.Error.Printf("failed to update config file: %v\n", err)
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
		configFile = "setup/configs/evm-config.yaml"
	case consts.WASM_ROLLAPP:
		configFile = "setup/configs/wasm-config.yaml"
	default:
		return fmt.Errorf("unsupported rollapp type: %s", rollappType)
	}

	data, err := configFiles.ReadFile(configFile)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	configPath := filepath.Join(destDir, "config.yaml")
	if err := os.MkdirAll(filepath.Dir(configPath), 0o755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	if err := os.WriteFile(configPath, data, 0o644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}
