package setup

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	cosmossdkmath "cosmossdk.io/math"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"

	initconfig "github.com/dymensionxyz/roller/cmd/config/init"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/dependencies"
	"github.com/dymensionxyz/roller/utils/filesystem"
	"github.com/dymensionxyz/roller/utils/rollapp"
	"github.com/dymensionxyz/roller/utils/roller"
)

//go:embed configs/*
var configFiles embed.FS

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "setup",
		Short: "Deploys an oracle to the RollApp",
		Run: func(cmd *cobra.Command, args []string) {
			if err := initconfig.AddFlags(cmd); err != nil {
				pterm.Error.Printf("failed to add flags: %v\n", err)
				return
			}

			if runtime.GOOS == "darwin" {
				pterm.Error.Println("darwin is not supported")
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

			cd := filepath.Join(rollerData.Home, consts.ConfigDirName.Oracle)
			oracle := NewOracle(rollerData)
			err = oracle.SetKey(rollerData)
			if err != nil {
				pterm.Error.Printf("failed to set oracle key: %v\n", err)
				return
			}

			codeID, err := oracle.GetCodeID()
			if err != nil {
				pterm.Error.Printf("failed to get code ID: %v\n", err)
				return
			}

			if codeID == "" {
				pterm.Info.Println("no code ID found, storing contract on chain")

				if err := oracle.StoreContract(rollerData); err != nil {
					pterm.Error.Printf("failed to store contract: %v\n", err)
					return
				}

				time.Sleep(time.Second * 2)

				codeID, err = oracle.GetCodeID()
				if err != nil {
					pterm.Error.Printf("failed to get code ID: %v\n", err)
					return
				}
			}

			oracle.CodeID = codeID
			raData, err := rollapp.GetMetadataFromChain(rollerData.RollappID, rollerData.HubData)
			if err != nil {
				pterm.Error.Printf("failed to get rollapp metadata: %v\n", err)
				return
			}

			pterm.Info.Printfln("code ID: %s", oracle.CodeID)

			pterm.Info.Println("checking for existing contracts...")

			contracts, err := oracle.ListContracts(rollerData)
			if err != nil {
				pterm.Error.Printf("failed to list contracts: %v\n", err)
				return
			}

			if len(contracts) > 0 {
				pterm.Info.Printfln("found existing contract: %s", contracts[0])
				oracle.ContractAddress = contracts[0]
			} else {
				pterm.Info.Println("no existing contracts found, instantiating contract...")
				if err := oracle.InstantiateContract(rollerData); err != nil {
					pterm.Error.Printf("failed to instantiate contract: %v\n", err)
					return
				}
			}

			pterm.Success.Println("oracle deployed successfully to chain")
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
				v = obvi.EvmOracle
			case consts.WASM_ROLLAPP:
				v = obvi.WasmOracle
			default:
				pterm.Error.Printfln("unsupported rollapp type %s", rollerData.RollappVMType)
				return
			}

			bc := dependencies.BinaryInstallConfig{
				RollappType: rollerData.RollappVMType,
				Version:     v,
				InstallDir:  consts.Executables.Oracle,
			}

			j, _ := json.MarshalIndent(bc, "", "  ")
			pterm.Info.Printfln("installing oracle binary:\n%s", string(j))

			err = dependencies.InstallBinary(context.Background(), bc)
			if err != nil {
				pterm.Error.Printf("failed to install oracle binary: %v\n", err)
				return
			}

			pterm.Info.Printfln("copying config file into %s", cd)
			if err := copyConfigFile(rollerData.RollappVMType, cd); err != nil {
				pterm.Error.Printf("failed to copy config file: %v\n", err)
				return
			}
			pterm.Info.Println("config file copied successfully")
			pterm.Info.Println("updating config values")
			gl, _ := cosmossdkmath.NewIntFromString(
				consts.DefaultMinGasPrice,
			)

			updates := map[string]any{
				"chainClient.oracleContractAddress": oracle.ContractAddress,
				"chainClient.fee":                   consts.DefaultTxFee,
				"chainClient.gasLimit":              gl,
				"chainClient.bech32Prefix":          raData.Rollapp.GenesisInfo.Bech32Prefix,
				"chainClient.chainId":               raData.Rollapp.RollappId,
				"chainClient.privateKey":            "oracle",
				"chainClient.ssl":                   false,
				"chainClient.chainGrpcHost":         "http://localhost:9090",
				"grpc_port":                         9093,
			}

			cfp := filepath.Join(cd, "config.yaml")
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
		configFile = "configs/evm-config.yaml"
	case consts.WASM_ROLLAPP:
		configFile = "configs/wasm-config.yaml"
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
