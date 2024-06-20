package initconfig

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
)

type RelayerFileChainConfig struct {
	Type  string                      `json:"type"`
	Value RelayerFileChainConfigValue `json:"value"`
}

type RelayerFileChainConfigValue struct {
	Key            string   `json:"key"`
	ChainID        string   `json:"chain-id"`
	RpcAddr        string   `json:"rpc-addr"`
	AccountPrefix  string   `json:"account-prefix"`
	KeyringBackend string   `json:"keyring-backend"`
	GasAdjustment  float64  `json:"gas-adjustment"`
	GasPrices      string   `json:"gas-prices"`
	Debug          bool     `json:"debug"`
	Timeout        string   `json:"timeout"`
	OutputFormat   string   `json:"output-format"`
	SignMode       string   `json:"sign-mode"`
	ExtraCodecs    []string `json:"extra-codecs"`
}

type RelayerChainConfig struct {
	ChainConfig relayer.ChainConfig
	GasPrices   string
	KeyName     string
}

func writeTmpChainConfig(chainConfig RelayerFileChainConfig, fileName string) (string, error) {
	file, err := json.Marshal(chainConfig)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(os.TempDir(), fileName)
	// nolint:gofumpt
	if err := os.WriteFile(filePath, file, 0644); err != nil {
		return "", err
	}
	return filePath, nil
}

func getRelayerFileChainConfig(relayerChainConfig RelayerChainConfig) RelayerFileChainConfig {
	return RelayerFileChainConfig{
		Type: "cosmos",
		Value: RelayerFileChainConfigValue{
			Key:            relayerChainConfig.KeyName,
			ChainID:        relayerChainConfig.ChainConfig.ID,
			RpcAddr:        relayerChainConfig.ChainConfig.RPC,
			AccountPrefix:  relayerChainConfig.ChainConfig.AddressPrefix,
			KeyringBackend: "test",
			GasAdjustment:  1.2,
			GasPrices:      relayerChainConfig.GasPrices,
			Debug:          true,
			Timeout:        "10s",
			OutputFormat:   "json",
			SignMode:       "direct",
			ExtraCodecs:    []string{"ethermint"},
		},
	}
}

func addChainToRelayer(fileChainConfig RelayerFileChainConfig, relayerHome string) error {
	chainFilePath, err := writeTmpChainConfig(fileChainConfig, "chain.json")
	if err != nil {
		return err
	}
	addChainCmd := exec.Command(
		consts.Executables.Relayer,
		"chains",
		"add",
		fileChainConfig.Value.ChainID,
		"--home",
		relayerHome,
		"--file",
		chainFilePath,
	)
	if err := addChainCmd.Run(); err != nil {
		return err
	}
	return nil
}

func initRelayer(relayerHome string) error {
	initRelayerConfigCmd := exec.Command(
		consts.Executables.Relayer,
		"config",
		"init",
		"--home",
		relayerHome,
	)
	return initRelayerConfigCmd.Run()
}

func addChainsConfig(
	rollappConfig relayer.ChainConfig,
	hubConfig relayer.ChainConfig,
	relayerHome string,
) error {
	relayerRollappConfig := getRelayerFileChainConfig(RelayerChainConfig{
		ChainConfig: rollappConfig,
		GasPrices:   rollappConfig.GasPrices + rollappConfig.Denom,
		KeyName:     consts.KeysIds.RollappRelayer,
	})

	relayerHubConfig := getRelayerFileChainConfig(RelayerChainConfig{
		ChainConfig: hubConfig,
		GasPrices:   hubConfig.GasPrices + hubConfig.Denom,
		KeyName:     consts.KeysIds.HubRelayer,
	})

	if err := addChainToRelayer(relayerRollappConfig, relayerHome); err != nil {
		return err
	}
	if err := addChainToRelayer(relayerHubConfig, relayerHome); err != nil {
		return err
	}
	return nil
}

func initializeRelayerConfig(
	rollappConfig relayer.ChainConfig,
	hubConfig relayer.ChainConfig,
	initConfig config.RollappConfig,
) error {
	relayerHome := filepath.Join(initConfig.Home, consts.ConfigDirName.Relayer)
	if err := initRelayer(relayerHome); err != nil {
		return err
	}
	if err := addChainsConfig(rollappConfig, hubConfig, relayerHome); err != nil {
		return err
	}
	if err := relayer.CreatePath(initConfig); err != nil {
		return err
	}
	return nil
}
