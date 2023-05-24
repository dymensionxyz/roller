package cmd

import (
	"os"
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
	"github.com/spf13/cobra"
)

const hubRPC string = "https://rpc-hub-35c.dymension.xyz:443"
const lightNodeEndpointFlag = "light-node-endpoint"

func createKey(relativePath string, keyId string, coinType ...uint32) (keyring.Info, error) {
	if len(coinType) == 0 {
		const cosmosDefaultCointype uint32 = 118
		coinType = []uint32{cosmosDefaultCointype}
	}
	rollappAppName := "rollapp"
	kr, err := keyring.New(
		rollappAppName,
		keyring.BackendTest,
		filepath.Join(os.Getenv("HOME"), relativePath),
		nil,
	)
	if err != nil {
		return nil, err
	}
	bip44Params := hd.NewFundraiserParams(0, coinType[0], 0)
	info, _, err := kr.NewMnemonic(keyId, keyring.English, bip44Params.String(), "", hd.Secp256k1)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func generateRollappKeys() (keyring.Info, error) {
	return createKey(".rollapp", "rollapp_sequencer")
}

var initCmd = &cobra.Command{
	Use:   "init <chain-id>",
	Short: "Initialize a rollapp configuration on your local machine",
	Run: func(cmd *cobra.Command, args []string) {
		const evmCoinType uint32 = 60
		rollappConfigDir := ".rollapp"
		relayerConfigDir := ".relayer"
		hubChainId := "internal-devnet"
		relayerKeysDirName := "keys"
		createKey(rollappConfigDir, "hub_sequencer")
		createKey(rollappConfigDir, "rollapp_sequencer", evmCoinType)
		relayerRollappDir := path.Join(relayerConfigDir, relayerKeysDirName, args[0])
		relayerHubDir := path.Join(relayerConfigDir, relayerKeysDirName, hubChainId)
		createKey(relayerHubDir, "relayer-hub-key")
		createKey(relayerRollappDir, "relayer-rollapp-key", evmCoinType)
		if !cmd.Flags().Changed(lightNodeEndpointFlag) {
			createKey(".light_node", "my-celes-key")
		}
	},
	Args: cobra.ExactArgs(1),
}

func init() {
	configCmd.AddCommand(initCmd)
	initCmd.Flags().StringP("hub-rpc", "", hubRPC, "Dymension Hub rpc endpoint")
	initCmd.Flags().StringP(lightNodeEndpointFlag, "", "", "The data availability light node endpoint. Runs an Arabica Celestia light node if not provided.")
	initCmd.Flags().StringP("denom", "", "", "The rollapp token smallest denominator, for example `wei` in Ethereum.")
	initCmd.Flags().StringP("key-prefix", "", "", "The `bech32` prefix of the rollapp keys.")
	initCmd.Flags().StringP("rollapp-binary", "", "", "The rollapp binary. Should be passed only if you built a custom rollapp.")
	initCmd.Flags().Int64P("decimals", "", 18, "The number of decimal places a rollapp token supports.")
}
