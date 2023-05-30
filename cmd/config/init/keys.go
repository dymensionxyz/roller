package init

import (
	"os"
	"path"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keyring"
)

func createKey(relativePath string, keyId string, coinType ...uint32) (keyring.Info, error) {
	var coinTypeVal = cosmosDefaultCointype
	if len(coinType) != 0 {
		coinTypeVal = coinType[0]
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
	bip44Params := hd.NewFundraiserParams(0, coinTypeVal, 0)
	info, _, err := kr.NewMnemonic(keyId, keyring.English, bip44Params.String(), "", hd.Secp256k1)
	if err != nil {
		return nil, err
	}
	return info, nil
}

func generateKeys(createLightNode bool, chainId string) {
	createKey(rollappConfigDir, keyNames.HubSequencer)
	createKey(rollappConfigDir, keyNames.RollappSequencer, evmCoinType)
	relayerRollappDir := path.Join(relayerConfigDir, relayerKeysDirName, chainId)
	relayerHubDir := path.Join(relayerConfigDir, relayerKeysDirName, hubChainId)
	createKey(relayerHubDir, "relayer-hub-key")
	createKey(relayerRollappDir, keyNames.RollappRelayer, evmCoinType)
	if createLightNode {
		createKey(".light_node", "my-celes-key")
	}
}
