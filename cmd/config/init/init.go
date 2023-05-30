package init

import (
	"os"
	"os/exec"
	"path/filepath"

	toml "github.com/pelletier/go-toml"
	"github.com/spf13/cobra"
)

var flagNames = struct {
	LightNodeEndpoint string
	Denom             string
	KeyPrefix         string
	Decimals          string
	RollappBinary     string
	HubRPC            string
}{
	LightNodeEndpoint: "light-node-endpoint",
	Denom:             "denom",
	KeyPrefix:         "key-prefix",
	Decimals:          "decimals",
	RollappBinary:     "rollapp-binary",
	HubRPC:            "hub-rpc",
}

var keyNames = struct {
	HubSequencer     string
	RollappSequencer string
	RollappRelayer   string
}{
	HubSequencer:     "hub_sequencer",
	RollappSequencer: "rollapp_sequencer",
	RollappRelayer:   "relayer-rollapp-key",
}

const hubRPC = "https://rpc-hub-35c.dymension.xyz:443"
const lightNodeEndpointFlag = "light-node-endpoint"

const evmCoinType uint32 = 60
const rollappConfigDir = ".rollapp"
const relayerConfigDir = ".relayer"
const hubChainId = "35-C"
const relayerKeysDirName = "keys"
const cosmosDefaultCointype uint32 = 118

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <chain-id> <denom>",
		Short: "Initialize a rollapp configuration on your local machine",
		Run: func(cmd *cobra.Command, args []string) {
			chainId := args[0]
			denom := args[1]
			rollappBinaryPath := getRollappBinaryPath(cmd.Flag(flagNames.RollappBinary).Value.String())
			decimals, err := cmd.Flags().GetUint64(flagNames.Decimals)
			if err != nil {
				panic(err)
			}
			generateKeys(cmd.Flags().Changed(lightNodeEndpointFlag), chainId)
			initializeRollappConfig(rollappBinaryPath, chainId, denom)
			err = initializeRollappGenesis(rollappBinaryPath, decimals, denom)
			if err != nil {
				panic(err)
			}
		},
		Args: cobra.ExactArgs(2),
	}
	cmd.Flags().StringP(flagNames.HubRPC, "", hubRPC, "Dymension Hub rpc endpoint")
	cmd.Flags().StringP(flagNames.LightNodeEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided.")
	cmd.Flags().StringP("key-prefix", "", "", "The `bech32` prefix of the rollapp keys.")
	cmd.Flags().StringP("rollapp-binary", "", "", "The rollapp binary. Should be passed only if you built a custom rollapp.")
	cmd.Flags().Uint64P(flagNames.Decimals, "", 18, "The number of decimal places a rollapp token supports.")
	return cmd
}

func getRollappBinaryPath(rollappBinaryPath string) string {
	if rollappBinaryPath == "" {
		rollappBinaryPath = "/usr/local/bin/rollapp_evm"
	}
	return rollappBinaryPath
}

func setRollappAppConfig(appConfigFilePath string, denom string) {
	config, _ := toml.LoadFile(appConfigFilePath)
	config.Set("minimum-gas-prices", "0"+denom)
	config.Set("api.enable", "true")
	config.Set("grpc.address", "0.0.0.0:8080")
	config.Set("grpc-web.address", "0.0.0.0:8081")
	file, _ := os.Create(appConfigFilePath)
	file.WriteString(config.String())
	file.Close()
}

func initializeRollappConfig(rollappExecutablePath string, chainId string, denom string) {
	initRollappCmd := exec.Command(rollappExecutablePath, "init", keyNames.HubSequencer, "--chain-id", chainId, "--home", filepath.Join(os.Getenv("HOME"), rollappConfigDir))
	err := initRollappCmd.Run()
	if err != nil {
		panic(err)
	}
	setRollappAppConfig(filepath.Join(os.Getenv("HOME"), rollappConfigDir, "config/app.toml"), denom)
}
