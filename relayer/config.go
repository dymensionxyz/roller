package relayer

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"gopkg.in/yaml.v2"
	"os"
	"os/exec"
	"path/filepath"
)

func CreatePath(rlpCfg config.RollappConfig) error {
	relayerHome := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer)
	setSettlementCmd := exec.Command(consts.Executables.Relayer, "chains", "set-settlement",
		rlpCfg.HubData.ID, "--home", relayerHome)
	if err := setSettlementCmd.Run(); err != nil {
		return err
	}
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
	if err := SetNestedValue(rlyCfg, keyPath, newValue); err != nil {
		return err
	}
	newData, err := yaml.Marshal(rlyCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %v", err)
	}
	return os.WriteFile(rlyConfigPath, newData, 0644)
}

func SetNestedValue(data map[interface{}]interface{}, keyPath []string, value interface{}) error {
	if len(keyPath) == 0 {
		return fmt.Errorf("empty key path")
	}
	if len(keyPath) == 1 {
		if value == nil {
			delete(data, keyPath[0])
		} else {
			data[keyPath[0]] = value
		}
		return nil
	}
	nextMap, ok := data[keyPath[0]].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("failed to get nested map for key: %s", keyPath[0])
	}
	return SetNestedValue(nextMap, keyPath[1:], value)
}

func GetNestedValue(data map[interface{}]interface{}, keyPath []string) (interface{}, error) {
	if len(keyPath) == 0 {
		return nil, fmt.Errorf("empty key path")
	}
	value, ok := data[keyPath[0]]
	if !ok {
		return nil, fmt.Errorf("key not found: %s", keyPath[0])
	}
	if len(keyPath) == 1 {
		return value, nil
	}
	nextMap, ok := value.(map[interface{}]interface{})
	if !ok {
		return nil, fmt.Errorf("failed to get nested map for key: %s", keyPath[0])
	}
	return GetNestedValue(nextMap, keyPath[1:])
}

func ReadRlyConfigValue(rlpCfg config.RollappConfig, keyPath []string) (interface{}, error) {
	rlyConfigPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer, "config", "config.yaml")
	data, err := os.ReadFile(rlyConfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load %s: %v", rlyConfigPath, err)
	}
	var rlyCfg map[interface{}]interface{}
	err = yaml.Unmarshal(data, &rlyCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal yaml: %v", err)
	}
	return GetNestedValue(rlyCfg, keyPath)
}

func ReadRlyConfig(rlpCfg config.RollappConfig) (map[interface{}]interface{}, error) {
	rlyConfigPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer, "config", "config.yaml")
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

func WriteRlyConfig(rlpCfg config.RollappConfig, rlyCfg map[interface{}]interface{}) error {
	rlyConfigPath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Relayer, "config", "config.yaml")
	data, err := yaml.Marshal(rlyCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %v", err)
	}
	return os.WriteFile(rlyConfigPath, data, 0644)
}
