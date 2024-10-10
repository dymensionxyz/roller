package set

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/dymensionxyz/roller/cmd/consts"
	"github.com/dymensionxyz/roller/relayer"
	"github.com/dymensionxyz/roller/utils/config/tomlconfig"
	"github.com/dymensionxyz/roller/utils/roller"
)

func setRollappRPC(rlpCfg roller.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	if err := relayer.UpdateRlyConfigValue(
		rlpCfg, []string{"chains", rlpCfg.RollappID, "value", "rpc-addr"}, "http://localhost:"+
			value,
	); err != nil {
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

func updateRlpClientCfg(rlpCfg roller.RollappConfig, newRpcPort string) error {
	configFilePath := filepath.Join(
		rlpCfg.Home,
		consts.ConfigDirName.Rollapp,
		"config",
		"client.toml",
	)
	return tomlconfig.UpdateFieldInFile(configFilePath, "node", "tcp://localhost:"+newRpcPort)
}

func updateRlpCfg(rlpCfg roller.RollappConfig, newRpc string) error {
	configFilePath := filepath.Join(
		rlpCfg.Home,
		consts.ConfigDirName.Rollapp,
		"config",
		"config.toml",
	)
	return tomlconfig.UpdateFieldInFile(configFilePath, "rpc.laddr", "tcp://0.0.0.0:"+newRpc)
}

func setJsonRpcPort(cfg roller.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	appCfgFilePath := filepath.Join(cfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	return tomlconfig.UpdateFieldInFile(appCfgFilePath, "json-rpc.address", "0.0.0.0:"+value)
}

func setWSPort(cfg roller.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	appCfgFilePath := filepath.Join(cfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	return tomlconfig.UpdateFieldInFile(appCfgFilePath, "json-rpc.ws-address", "0.0.0.0:"+value)
}

func setGRPCPort(cfg roller.RollappConfig, value string) error {
	if err := validatePort(value); err != nil {
		return err
	}
	appCfgFilePath := filepath.Join(cfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	return tomlconfig.UpdateFieldInFile(appCfgFilePath, "grpc.address", "0.0.0.0:"+value)
}
