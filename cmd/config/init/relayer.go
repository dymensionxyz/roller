package initconfig

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
)

type RelayerFileChainConfig struct {
	Type  string                      `json:"type"`
	Value RelayerFileChainConfigValue `json:"value"`
}
type RelayerFileChainConfigValue struct {
	Key            string  `json:"key"`
	ChainID        string  `json:"chain-id"`
	RpcAddr        string  `json:"rpc-addr"`
	AccountPrefix  string  `json:"account-prefix"`
	KeyringBackend string  `json:"keyring-backend"`
	GasAdjustment  float64 `json:"gas-adjustment"`
	GasPrices      string  `json:"gas-prices"`
	Debug          bool    `json:"debug"`
	Timeout        string  `json:"timeout"`
	OutputFormat   string  `json:"output-format"`
	SignMode       string  `json:"sign-mode"`
	ClientType     string  `json:"client-type"`
}

type ChainConfig struct {
	ID            string
	RPC           string
	Denom         string
	AddressPrefix string
}

type RelayerChainConfig struct {
	ChainConfig ChainConfig
	GasPrices   string
	ClientType  string
	KeyName     string
}

func writeTmpChainConfig(chainConfig RelayerFileChainConfig, fileName string) (string, error) {
	file, err := json.Marshal(chainConfig)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(os.TempDir(), fileName)
	if err := ioutil.WriteFile(filePath, file, 0644); err != nil {
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
			ClientType:     relayerChainConfig.ClientType,
		},
	}
}

func addChainToRelayer(fileChainConfig RelayerFileChainConfig, relayerHome string) error {
	chainFilePath, err := writeTmpChainConfig(fileChainConfig, "chain.json")
	if err != nil {
		return err
	}
	addChainCmd := exec.Command(consts.Executables.Relayer, "chains", "add", fileChainConfig.Value.ChainID, "--home", relayerHome, "--file", chainFilePath)
	if err := addChainCmd.Run(); err != nil {
		return err
	}
	return nil
}

func initRelayer(relayerHome string) error {
	initRelayerConfigCmd := exec.Command(consts.Executables.Relayer, "config", "init", "--home", relayerHome)
	return initRelayerConfigCmd.Run()
}

func addChainsConfig(rollappConfig ChainConfig, hubConfig ChainConfig, relayerHome string) error {
	relayerRollappConfig := getRelayerFileChainConfig(RelayerChainConfig{
		ChainConfig: rollappConfig,
		GasPrices:   "0.0" + rollappConfig.Denom,
		ClientType:  "01-dymint",
		KeyName:     consts.KeyNames.RollappRelayer,
	})

	relayerHubConfig := getRelayerFileChainConfig(RelayerChainConfig{
		ChainConfig: hubConfig,
		GasPrices:   "0.25" + hubConfig.Denom,
		ClientType:  "07-tendermint",
		KeyName:     consts.KeyNames.HubRelayer,
	})

	if err := addChainToRelayer(relayerRollappConfig, relayerHome); err != nil {
		return err
	}
	if err := addChainToRelayer(relayerHubConfig, relayerHome); err != nil {
		return err
	}
	return nil
}

func setupPath(rollappConfig ChainConfig, hubConfig ChainConfig, relayerHome string) error {
	setSettlementCmd := exec.Command(consts.Executables.Relayer, "chains", "set-settlement", hubConfig.ID, "--home", relayerHome)
	if err := setSettlementCmd.Run(); err != nil {
		return err
	}
	relayerPath := "hub-rollapp"
	newPathCmd := exec.Command(consts.Executables.Relayer, "paths", "new", rollappConfig.ID, hubConfig.ID, relayerPath, "--src-port", "transfer", "--dst-port", "transfer", "--version", "ics20-1", "--home", relayerHome)
	if err := newPathCmd.Run(); err != nil {
		return err
	}
	return nil
}

func initializeRelayerConfig(rollappConfig ChainConfig, hubConfig ChainConfig, initConfig InitConfig) error {
	relayerHome := filepath.Join(initConfig.Home, ConfigDirName.Relayer)
	if err := initRelayer(relayerHome); err != nil {
		return err
	}
	if err := addChainsConfig(rollappConfig, hubConfig, relayerHome); err != nil {
		return err
	}
	if err := setupPath(rollappConfig, hubConfig, relayerHome); err != nil {
		return err
	}
	return nil
}
