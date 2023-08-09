package set

import (
	"fmt"
	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"
	"github.com/pelletier/go-toml"
	"gopkg.in/yaml.v2"
	"os"
	"path/filepath"
	"strconv"
)

func setRollappRPC(rlpCfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if err := updateRlyConfigValue(rlpCfg, []string{"chains", rlpCfg.RollappID, "value", "rpc-addr"}, "http://localhost:"+
		value); err != nil {
		return err
	}
	if err := updateRlpCfg(rlpCfg, value); err != nil {
		return err
	}
	return updateRlpClientCfg(rlpCfg, value)
}

func validatePort(portStr string) error {
	_, err := strconv.Atoi(portStr)
	if err != nil {
		return fmt.Errorf("port should be a number: %s", portStr)
	}
	return nil
}

func updateRlyConfigValue(rlpCfg config.RollappConfig, keyPath []string, newValue string) error {
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
	if err := setNestedValue(rlyCfg, keyPath, newValue); err != nil {
		return err
	}
	newData, err := yaml.Marshal(rlyCfg)
	if err != nil {
		return fmt.Errorf("failed to marshal updated config: %v", err)
	}
	return os.WriteFile(rlyConfigPath, newData, 0644)
}

func setNestedValue(data map[interface{}]interface{}, keyPath []string, value string) error {
	if len(keyPath) == 0 {
		return fmt.Errorf("empty key path")
	}
	if len(keyPath) == 1 {
		data[keyPath[0]] = value
		return nil
	}
	nextMap, ok := data[keyPath[0]].(map[interface{}]interface{})
	if !ok {
		return fmt.Errorf("failed to get nested map for key: %s", keyPath[0])
	}
	return setNestedValue(nextMap, keyPath[1:], value)
}

func updateRlpClientCfg(rlpCfg config.RollappConfig, newRpcPort string) error {
	configFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "client.toml")
	return updateFieldInToml(configFilePath, "node", "tcp://localhost:"+newRpcPort)
}

func updateRlpCfg(rlpCfg config.RollappConfig, newRpc string) error {
	configFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "config.toml")
	return updateFieldInToml(configFilePath, "rpc.laddr", "tcp://0.0.0.0:"+newRpc)
}

func updateFieldInToml(tmlFilePath, key, value string) error {
	var tomlCfg, err = toml.LoadFile(tmlFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", tmlFilePath, err)
	}
	tomlCfg.Set(key, value)
	return utils.WriteTomlTreeToFile(tomlCfg, tmlFilePath)
}
