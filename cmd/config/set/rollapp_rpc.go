package set

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/config"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils"
)

func setRollappRPC(rlpCfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if err := relayer.UpdateRlyConfigValue(rlpCfg, []string{"chains", rlpCfg.RollappID, "value", "rpc-addr"}, "http://localhost:"+
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

func updateRlpClientCfg(rlpCfg config.RollappConfig, newRpcPort string) error {
	configFilePath := filepath.Join(
		rlpCfg.Home,
		consts.ConfigDirName.Rollapp,
		"config",
		"client.toml",
	)
	return utils.UpdateFieldInToml(configFilePath, "node", "tcp://localhost:"+newRpcPort)
}

func updateRlpCfg(rlpCfg config.RollappConfig, newRpc string) error {
	configFilePath := filepath.Join(
		rlpCfg.Home,
		consts.ConfigDirName.Rollapp,
		"config",
		"config.toml",
	)
	return utils.UpdateFieldInToml(configFilePath, "rpc.laddr", "tcp://0.0.0.0:"+newRpc)
}

func setJsonRpcPort(cfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	appCfgFilePath := filepath.Join(cfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	return utils.UpdateFieldInToml(appCfgFilePath, "json-rpc.address", "0.0.0.0:"+value)
}

func setWSPort(cfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	appCfgFilePath := filepath.Join(cfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	return utils.UpdateFieldInToml(appCfgFilePath, "json-rpc.ws-address", "0.0.0.0:"+value)
}

func setGRPCPort(cfg config.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	appCfgFilePath := filepath.Join(cfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	return utils.UpdateFieldInToml(appCfgFilePath, "grpc.address", "0.0.0.0:"+value)
}
