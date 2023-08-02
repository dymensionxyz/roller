package initconfig

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/sequencer"

	"github.com/dymensionxyz/roller/cmd/utils"
	"github.com/dymensionxyz/roller/config"

	"github.com/dymensionxyz/roller/cmd/consts"
	toml "github.com/pelletier/go-toml"
)

func initializeRollappConfig(initConfig config.RollappConfig) error {
	home := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)

	initRollappCmd := exec.Command(initConfig.RollappBinary, "init", consts.KeysIds.HubSequencer, "--chain-id",
		initConfig.RollappID, "--home", home)
	_, err := utils.ExecBashCommandWithStdout(initRollappCmd)
	if err != nil {
		return err
	}

	seqPubKey, err := utils.GetSequencerPubKey(initConfig)
	if err != nil {
		return err
	}

	setGentxCmd := exec.Command(initConfig.RollappBinary, "gentx_seq",
		"--pubkey", seqPubKey, "--from", consts.KeysIds.RollappSequencer, "--keyring-backend", "test", "--home", home)
	_, err = utils.ExecBashCommandWithStdout(setGentxCmd)
	if err != nil {
		return err
	}

	err = setRollappAppConfig(filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp, "config", "app.toml"),
		initConfig.Denom)
	if err != nil {
		return err
	}

	setConfigCmd = exec.Command(initConfig.RollappBinary, "config", "keyring-backend", "os", "--home", home)
	_, err = utils.ExecBashCommandWithStdout(setConfigCmd)
	if err != nil {
		return err
	}
	if err = sequencer.SetDefaultDymintConfig(initConfig); err != nil {
		return err
	}
	return nil
}

func setRollappConfig(rlpCfg config.RollappConfig) error {
	if err := setAppConfig(rlpCfg); err != nil {
		return err
	}
	if err := setTMConfig(rlpCfg); err != nil {
		return err
	}
	if err := sequencer.SetDefaultDymintConfig(rlpCfg); err != nil {
		return err
	}
	return nil
}

func setAppConfig(rlpCfg config.RollappConfig) error {
	appConfigFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "app.toml")
	appCfg, err := toml.LoadFile(appConfigFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", appConfigFilePath, err)
	}

	appCfg.Set("minimum-gas-prices", "0"+rlpCfg.Denom)
	appCfg.Set("api.enable", "true")

	if appCfg.Has("json-rpc") {
		appCfg.Set("json-rpc.address", "0.0.0.0:8545")
		appCfg.Set("json-rpc.ws-address", "0.0.0.0:8546")
	}
	return writeTomlTreeToFile(appCfg, appConfigFilePath)
}

func setTMConfig(rlpCfg config.RollappConfig) error {
	configFilePath := filepath.Join(rlpCfg.Home, consts.ConfigDirName.Rollapp, "config", "config.toml")
	var tomlCfg, err = toml.LoadFile(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load %s: %v", configFilePath, err)
	}
	tomlCfg.Set("rpc.laddr", "tcp://0.0.0.0:26657")
	return writeTomlTreeToFile(tomlCfg, configFilePath)
}

func writeTomlTreeToFile(tomlConfig *toml.Tree, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	_, err = file.WriteString(tomlConfig.String())
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}
	return nil
}

func RollappConfigDir(root string) string {
	return filepath.Join(root, consts.ConfigDirName.Rollapp, "config")
}
