package init

import (
	"os"
	"os/exec"
	"path/filepath"

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

const hubRPC = "https://rpc-hub-35c.dymension.xyz:443"
const lightNodeEndpointFlag = "light-node-endpoint"
const hubSequencerKeyName = "hub_sequencer"

const evmCoinType uint32 = 60
const rollappConfigDir string = ".rollapp"
const relayerConfigDir string = ".relayer"
const hubChainId string = "internal-devnet"
const relayerKeysDirName string = "keys"
const cosmosDefaultCointype uint32 = 118
const sequencerKeyName string = "rollapp_sequencer"

func getDenom(denom string, chainId string) string {
	if denom == "" {
		return "a" + chainId[:3]
	}
	return denom
}

func getRollappBinaryPath(rollappBinaryPath string) string {
	if rollappBinaryPath == "" {
		rollappBinaryPath = "/usr/local/bin/rollapp_evm"
	}
	return rollappBinaryPath
}

func initializeRollappConfig(rollappExecutablePath string, chainId string, denom string) {
	initRollappCmd := exec.Command(rollappExecutablePath, "init", hubSequencerKeyName, "--chain-id", chainId, "--home", filepath.Join(os.Getenv("HOME"), rollappConfigDir))
	err := initRollappCmd.Run()
	if err != nil {
		panic(err)
	}
	setRollappAppConfig(filepath.Join(os.Getenv("HOME"), rollappConfigDir, "config/app.toml"), denom)
}

func InitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init <chain-id>",
		Short: "Initialize a rollapp configuration on your local machine",
		Run: func(cmd *cobra.Command, args []string) {
			chainId := args[0]
			rollappBinaryPath := getRollappBinaryPath(cmd.Flag(flagNames.RollappBinary).Value.String())
			denom := getDenom(cmd.Flag(flagNames.Denom).Value.String(), chainId)
			generateKeys(cmd.Flags().Changed(lightNodeEndpointFlag), chainId)
			initializeRollappConfig(rollappBinaryPath, chainId, denom)
		},
		Args: cobra.ExactArgs(1),
	}
	cmd.Flags().StringP(flagNames.HubRPC, "", hubRPC, "Dymension Hub rpc endpoint")
	cmd.Flags().StringP(flagNames.LightNodeEndpoint, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided.")
	cmd.Flags().StringP("denom", "", "", "The rollapp token smallest denominator, for example `wei` in Ethereum.")
	cmd.Flags().StringP("key-prefix", "", "", "The `bech32` prefix of the rollapp keys.")
	cmd.Flags().StringP("rollapp-binary", "", "", "The rollapp binary. Should be passed only if you built a custom rollapp.")
	cmd.Flags().Int64P("decimals", "", 18, "The number of decimal places a rollapp token supports.")
	return cmd
}
