package relayer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/utils"
	"gopkg.in/yaml.v2"
)

func CreatePath(rlpCfg config.RollappConfig) error {
	relayerHome := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer)
	newPathCmd := exec.Command(consts.Executables.Relayer, "paths", "new", rlpCfg.HubData.ID, rlpCfg.RollappID,
		consts.DefaultRelayerPath, "--home", relayerHome)
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

func UpdateRlyConfigValue(rlpCfg config.RollappConfig, keyPath []string, newValue interface{}) error {
	rlyConfigPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer, "config", "config.yaml")
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
	return os.WriteFile(rlyConfigPath, newData, 0644)
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
	return os.WriteFile(rlyConfigPath, data, 0644)
}
