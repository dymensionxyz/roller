package relayer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pterm/pterm"
	yaml "gopkg.in/yaml.v3"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/utils"
	"github.com/dymensionxyz/roller/utils/config/yamlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
)

type RelayerFileChainConfig struct {
	Type  string                      `yaml:"type"  json:"type"`
	Value RelayerFileChainConfigValue `yaml:"value" json:"value"`
}

type RelayerFileChainConfigValue struct {
	Key            string   `yaml:"key"             json:"key"`
	ChainID        string   `yaml:"chain-id"        json:"chain-id"`
	RpcAddr        string   `yaml:"rpc-addr"        json:"rpc-addr"`
	ApiAddr        string   `yaml:"http-addr"       json:"http-addr"`
	AccountPrefix  string   `yaml:"account-prefix"  json:"account-prefix"`
	KeyringBackend string   `yaml:"keyring-backend" json:"keyring-backend"`
	GasAdjustment  float64  `yaml:"gas-adjustment"  json:"gas-adjustment"`
	GasPrices      string   `yaml:"gas-prices"      json:"gas-prices"`
	Debug          bool     `yaml:"debug"           json:"debug"`
	Timeout        string   `yaml:"timeout"         json:"timeout"`
	OutputFormat   string   `yaml:"output-format"   json:"output-format"`
	SignMode       string   `yaml:"sign-mode"       json:"sign-mode"`
	ExtraCodecs    []string `yaml:"extra-codecs"    json:"extra-codecs"`
}

type RelayerChainConfig struct {
	ChainConfig ChainConfig
	GasPrices   string
	KeyName     string
}

// Config struct represents the paths section inside the relayer
// configuration file
type Config struct {
	Chains map[string]RelayerFileChainConfig `yaml:"chains"`
	Paths  *Paths                            `yaml:"paths"`
}

type Paths struct {
	HubRollapp *struct {
		Dst *struct {
			ChainID      string `yaml:"chain-id"`
			ClientID     string `yaml:"client-id"`
			ConnectionID string `yaml:"connection-id"`
		} `yaml:"dst"`
		Src *struct {
			ChainID      string `yaml:"chain-id"`
			ClientID     string `yaml:"client-id"`
			ConnectionID string `yaml:"connection-id"`
		} `yaml:"src"`
		SrcChannelFilter *struct {
			ChannelList []string `yaml:"channel-list"`
			Rule        string   `yaml:"rule"`
		} `yaml:"src-channel-filter"`
	} `yaml:"hub-rollapp"`
}

func (c *Config) GetChains(cfgPath string) ([]string, error) {
	err := c.Load(cfgPath)
	if err != nil {
		return nil, err
	}

	chains := make([]string, 0, len(c.Chains))
	for chainName := range c.Chains {
		chains = append(chains, chainName)
	}
	return chains, nil
}

func (c *Config) Load(rlyConfigPath string) error {
	fmt.Println("loading config from", rlyConfigPath)
	data, err := os.ReadFile(rlyConfigPath)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(data, c)
	if err != nil {
		return err
	}

	return nil
}

func (c *Config) GetPath() *Paths {
	if c.Paths == nil || c.Paths.HubRollapp == nil {
		return nil
	}

	return c.Paths
}

func (c *Config) CreatePath(rlpCfg roller.RollappConfig) error {
	relayerHome := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer)
	pterm.Info.Printf("creating new ibc path from %s to %s\n", rlpCfg.HubData.ID, rlpCfg.RollappID)

	newPathCmd := exec.Command(
		consts.Executables.Relayer,
		"paths",
		"new",
		rlpCfg.HubData.ID,
		rlpCfg.RollappID,
		consts.DefaultRelayerPath,
		"--home",
		relayerHome,
	)
	if err := newPathCmd.Run(); err != nil {
		return err
	}

	return nil
}

func DeletePath(rlpCfg roller.RollappConfig) error {
	relayerHome := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer)
	pterm.Info.Printf("removing ibc path from %s to %s\n", rlpCfg.HubData.ID, rlpCfg.RollappID)

	newPathCmd := exec.Command(
		consts.Executables.Relayer,
		"paths",
		"delete",
		consts.DefaultRelayerPath,
		"--home",
		relayerHome,
	)
	if err := newPathCmd.Run(); err != nil {
		return err
	}

	return nil
}

type ChainConfig struct {
	ID            string
	RPC           string
	Denom         string
	AddressPrefix string
	GasPrices     string
}

func UpdateRlyConfigValue(
	rlpCfg roller.RollappConfig,
	keyPath []string,
	newValue interface{},
) error {
	rlyConfigPath := filepath.Join(
		rlpCfg.Home,
		consts.ConfigDirName.Relayer,
		"config",
		"config.yaml",
	)

	data, err := os.ReadFile(rlyConfigPath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", rlyConfigPath, err)
	}

	var rlyCfg map[interface{}]interface{}

	err = yaml.Unmarshal(data, &rlyCfg)
	if err != nil {
		return fmt.Errorf("failed to unmarshal yaml: %v", err)
	}

	if err := utils.SetNestedValue(rlyCfg, keyPath, newValue); err != nil {
		return err
	}

	newData, err := yaml.Marshal(rlyCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %v", err)
	}

	// nolint:gofumpt
	return os.WriteFile(rlyConfigPath, newData, 0o644)
}

func ReadRlyConfig(homeDir string) (map[interface{}]interface{}, error) {
	rlyConfigPath := filepath.Join(homeDir, consts.ConfigDirName.Relayer, "config", "config.yaml")
	data, err := os.ReadFile(rlyConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %v", rlyConfigPath, err)
	}
	var rlyCfg map[interface{}]interface{}
	err = yaml.Unmarshal(data, &rlyCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}
	return rlyCfg, nil
}

func WriteRlyConfig(homeDir string, rlyCfg map[interface{}]interface{}) error {
	rlyConfigPath := filepath.Join(homeDir, consts.ConfigDirName.Relayer, "config", "config.yaml")
	data, err := yaml.Marshal(rlyCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	// nolint:gofumpt
	return os.WriteFile(rlyConfigPath, data, 0o644)
}

func writeTmpChainConfig(chainConfig RelayerFileChainConfig, fileName string) (string, error) {
	file, err := json.Marshal(chainConfig)
	if err != nil {
		return "", err
	}
	filePath := filepath.Join(os.TempDir(), fileName)
	// nolint:gofumpt
	if err := os.WriteFile(filePath, file, 0o644); err != nil {
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
	rollappConfig ChainConfig,
	hubConfig ChainConfig,
	relayerHome string,
) error {
	relayerRollappConfig := getRelayerFileChainConfig(
		RelayerChainConfig{
			ChainConfig: rollappConfig,
			GasPrices:   rollappConfig.GasPrices + rollappConfig.Denom,
			KeyName:     consts.KeysIds.RollappRelayer,
		},
	)

	relayerHubConfig := getRelayerFileChainConfig(
		RelayerChainConfig{
			ChainConfig: hubConfig,
			GasPrices:   hubConfig.GasPrices + hubConfig.Denom,
			KeyName:     consts.KeysIds.HubRelayer,
		},
	)

	if err := addChainToRelayer(relayerRollappConfig, relayerHome); err != nil {
		return err
	}
	if err := addChainToRelayer(relayerHubConfig, relayerHome); err != nil {
		return err
	}
	return nil
}

func InitializeConfig(
	rollappConfig ChainConfig,
	hubConfig ChainConfig,
	home string,
) error {
	relayerHome := filepath.Join(home, consts.ConfigDirName.Relayer)

	if err := initRelayer(relayerHome); err != nil {
		return err
	}
	if err := addChainsConfig(rollappConfig, hubConfig, relayerHome); err != nil {
		return err
	}

	return nil
}

func (c *Config) RollappID() string {
	return c.Paths.HubRollapp.Dst.ChainID
}

func (c *Config) HubDataFromRelayerConfig() *consts.HubData {
	hd := consts.HubData{
		ID:     c.Paths.HubRollapp.Src.ChainID,
		RpcUrl: c.Chains[c.Paths.HubRollapp.Src.ChainID].Value.RpcAddr,
		ApiUrl: c.Chains[c.Paths.HubRollapp.Src.ChainID].Value.ApiAddr,
	}

	return &hd
}

func (c *Config) RaDataFromRelayerConfig() *consts.RollappData {
	raData := consts.RollappData{
		ID:     c.Paths.HubRollapp.Dst.ChainID,
		RpcUrl: c.Chains[c.Paths.HubRollapp.Dst.ChainID].Value.RpcAddr,
	}

	return &raData
}

func (r *Relayer) UpdateConfigWithDefaultValues(rollerData roller.RollappConfig) error {
	updates := map[string]interface{}{
		fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.HubData.ID): 2.0,
		fmt.Sprintf("chains.%s.value.coin-type", rollerData.HubData.ID):      60,
		fmt.Sprintf("chains.%s.value.gas-prices", rollerData.HubData.ID): fmt.Sprintf(
			"2000000000000%s",
			consts.Denoms.Hub,
		),
		fmt.Sprintf("chains.%s.value.gas-adjustment", rollerData.RollappID):  1.3,
		fmt.Sprintf("chains.%s.value.min-gas-amount", rollerData.HubData.ID): 0,
		fmt.Sprintf("chains.%s.value.max-gas-amount", rollerData.HubData.ID): 100_000_000,
		fmt.Sprintf("chains.%s.value.is-dym-hub", rollerData.HubData.ID):     true,
		fmt.Sprintf(
			"chains.%s.value.http-addr",
			rollerData.HubData.ID,
		): rollerData.HubData.ApiUrl,
		fmt.Sprintf("chains.%s.value.is-dym-rollapp", rollerData.RollappID): true,
		"extra-codecs": []string{
			"ethermint",
		},
	}
	err := yamlconfig.UpdateNestedYAML(r.ConfigFilePath, updates)
	if err != nil {
		pterm.Error.Printf("Error updating YAML: %v\n", err)
		return err
	}

	return nil
}
