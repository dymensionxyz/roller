package initconfig

import (
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dymensionxyz/roller/cmd/utils"

	"github.com/dymensionxyz/roller/cmd/consts"
	toml "github.com/pelletier/go-toml"
)

func initializeRollappConfig(initConfig utils.RollappConfig) error {
	home := filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp)

	initRollappCmd := exec.Command(initConfig.RollappBinary, "init", consts.KeysIds.HubSequencer, "--chain-id",
		initConfig.RollappID, "--home", home)
	_, err := utils.ExecBashCommand(initRollappCmd)
	if err != nil {
		return err
	}

	setConfigCmd := exec.Command(initConfig.RollappBinary, "config", "keyring-backend", "test", "--home", home)
	_, err = utils.ExecBashCommand(setConfigCmd)
	if err != nil {
		return err
	}

	seqPubKey, err := utils.GetSequencerPubKey(initConfig)
	if err != nil {
		return err
	}

	setGentxCmd := exec.Command(initConfig.RollappBinary, "gentx_seq",
		"--pubkey", seqPubKey, "--from", consts.KeysIds.RollappSequencer, "--home", home)
	_, err = utils.ExecBashCommand(setGentxCmd)
	if err != nil {
		return err
	}

	err = setRollappAppConfig(filepath.Join(initConfig.Home, consts.ConfigDirName.Rollapp, "config/app.toml"),
		initConfig.Denom)
	if err != nil {
		return err
	}
	return nil
}

func setRollappAppConfig(appConfigFilePath string, denom string) error {
	//FIXME: why error not checked?
	config, _ := toml.LoadFile(appConfigFilePath)
	config.Set("minimum-gas-prices", "0"+denom)
	config.Set("api.enable", "true")
	file, err := os.Create(appConfigFilePath)
	if err != nil {
		return err
	}
	_, err = file.WriteString(config.String())
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
